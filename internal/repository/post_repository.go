package repository

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/viewsharp/technopark-forum/internal/db"
	"github.com/viewsharp/technopark-forum/internal/domain"
)

type PostRepository struct {
	DB      Database
	Queries *db.Queries
}

var regexInvalidAuthor, _ = regexp.Compile(`^Key \(user_nn\)=\(([\w\.]+)\) is not present in table "users"\.$`)

func (r *PostRepository) AddByThreadSlug(ctx context.Context, posts []domain.Post, slug string) error {
	dbThread, err := r.Queries.GetThreadBySlug(ctx, pgtype.Text{String: slug, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrPostNotFoundThread
		}
		return fmt.Errorf("select thread: %w", err)
	}

	return r.add(ctx, posts, dbThread.ID, dbThread.ForumSlug)
}

func (r *PostRepository) AddByThreadId(ctx context.Context, posts []domain.Post, threadId int32) error {
	dbThread, err := r.Queries.GetThreadByID(ctx, threadId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrPostNotFoundThread
		}
		return fmt.Errorf("select thread: %w", err)
	}

	return r.add(ctx, posts, threadId, dbThread.ForumSlug)
}

func (r *PostRepository) add(ctx context.Context, posts []domain.Post, threadId int32, forumSlug string) error {
	// select parents

	parentIDMap := make(map[int64]struct{})
	for _, post := range posts {
		if post.Parent != nil {
			parentIDMap[*post.Parent] = struct{}{}
		}
	}

	parentIDs := slices.AppendSeq(make([]int64, 0, len(parentIDMap)), maps.Keys(parentIDMap))
	parents, err := r.Queries.ListByID(ctx, parentIDs)
	if err != nil {
		return fmt.Errorf("list parents by id: %w", err)
	}

	parentByID := make(map[int64]db.Post, len(parentIDs))
	for _, parent := range parents {
		parentByID[parent.ID] = parent
	}

	// insert posts

	postsParams := make([]db.CreatePostsParams, 0, len(posts))
	for _, post := range posts {
		var parentID pgtype.Int8
		var path []int64
		if post.Parent != nil {
			if parent, ok := parentByID[*post.Parent]; ok {
				if parent.ThreadID != threadId {
					return domain.ErrPostInvalidParent
				}

				parentID = pgtype.Int8{Int64: parent.ID, Valid: true}
				path = append(parent.Path, parent.ID)
			} else {
				return domain.ErrPostInvalidParent
			}
		}

		postsParams = append(postsParams, db.CreatePostsParams{
			Message:  *post.Message,
			ParentID: parentID,
			UserNn:   *post.Author,
			ThreadID: threadId,
			Path:     path,
		})
	}

	postsBatch := r.Queries.CreatePosts(ctx, postsParams)
	postsBatch.QueryRow(func(i int, post db.Post, batchErr error) {
		if errors.Is(batchErr, db.ErrBatchAlreadyClosed) {
			return
		}
		if batchErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(batchErr, &pgErr) && pgErr.Code == "23503" {
				err = domain.ErrPostNotFoundUser{Nickname: post.UserNn}
			} else {
				err = batchErr
			}

			postsBatch.Close()
			return
		}

		posts[i].Author = &post.UserNn
		posts[i].Created = &post.Created.Time
		posts[i].Id = &post.ID
		posts[i].Message = &post.Message
		posts[i].Parent = &post.ParentID.Int64
		posts[i].Thread = &post.ThreadID
		posts[i].Forum = &forumSlug
	})
	if err != nil {
		return fmt.Errorf("create posts: %w", err)
	}

	// insert forum users

	forumUsersParams := make([]db.CreateForumUserParams, 0, len(posts))
	for _, post := range posts {
		forumUsersParams = append(forumUsersParams, db.CreateForumUserParams{
			ForumSlug: forumSlug,
			Nickname:  *post.Author,
		})
	}

	forumUsersBatch := r.Queries.CreateForumUser(ctx, forumUsersParams)
	forumUsersBatch.Exec(func(i int, batchErr error) {
		if batchErr != nil && !errors.Is(batchErr, db.ErrBatchAlreadyClosed) {
			err = batchErr
		}
		forumUsersBatch.Close()
	})

	// update posts count

	err = r.Queries.IncreasePostsCount(ctx, db.IncreasePostsCountParams{
		NewPostsCount: int32(len(posts)),
		Slug:          forumSlug,
	})
	if err != nil {
		return fmt.Errorf("update forum: %w", err)
	}

	return nil
}

func (r *PostRepository) ById(ctx context.Context, id int64, related []string) (*domain.PostFull, error) {
	userObj := domain.User{}
	forumObj := domain.Forum{}
	postObj := domain.Post{}
	threadObj := domain.Thread{}
	result := domain.PostFull{}

	err := r.DB.QueryRow(
		ctx,
		`	SELECT 
					u.about, u.email, u.fullname, u.nickname, 
					f.posts, f.slug, f.threads, f.title, f.user_nn, 
					p.user_nn, p.created, f.slug, p.id, p.isedited, p.message, p.parent_id, p.thread_id,
					t.user_nn, t.created, f.slug, t.id, t.message, t.slug, t.title, t.votes
				FROM posts p
					JOIN users u ON p.user_nn = u.nickname
					JOIN threads t ON p.thread_id = t.id
					JOIN forums f ON t.forum_slug = f.slug
				WHERE p.id = $1`,
		id,
	).Scan(
		&userObj.About, &userObj.Email, &userObj.FullName, &userObj.Nickname,
		&forumObj.Posts, &forumObj.Slug, &forumObj.Threads, &forumObj.Title, &forumObj.User,
		&postObj.Author, &postObj.Created, &postObj.Forum, &postObj.Id, &postObj.IsEdited, &postObj.Message, &postObj.Parent, &postObj.Thread,
		&threadObj.Author, &threadObj.Created, &threadObj.Forum, &threadObj.Id, &threadObj.Message, &threadObj.Slug, &threadObj.Title, &threadObj.Votes,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get post: %w", err)
	}

	result.Post = &postObj
	for _, relate := range related {
		switch relate {
		case "user":
			result.Author = &userObj
		case "thread":
			result.Thread = &threadObj
		case "forum":
			result.Forum = &forumObj
		}
	}
	return &result, nil
}

func (r *PostRepository) UpdateById(ctx context.Context, id int64, post domain.PostUpdate) error {
	if post.Message == nil {
		return nil
	}

	_, err := r.DB.Exec(
		ctx,
		`	UPDATE posts 
				SET message = $1, isedited = TRUE
				WHERE id = $2`,
		post.Message, id,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("get post: %w", err)
	}

	return nil
}

func (r *PostRepository) FlatByThreadSlug(ctx context.Context, slug string, limit int32, desc bool, since int64) ([]domain.Post, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`	SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id
										FROM posts p
											JOIN threads t ON p.thread_id = t.id
										WHERE t.slug = $1`)

	if since != 0 {
		if desc {
			queryBuilder.WriteString(" AND p.id < $3")
		} else {
			queryBuilder.WriteString(" AND p.id > $3")
		}
	}

	if desc {
		queryBuilder.WriteString(" ORDER BY p.created DESC, p.id DESC LIMIT $2")
	} else {
		queryBuilder.WriteString(" ORDER BY p.created, p.id LIMIT $2")
	}

	return r.bySlug(ctx, queryBuilder.String(), slug, limit, since)
}

func (r *PostRepository) FlatByThreadId(ctx context.Context, id int, limit int32, desc bool, since int64) ([]domain.Post, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`	SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id
										FROM posts p
											JOIN threads t ON p.thread_id = t.id
										WHERE t.id = $1`)

	if since != 0 {
		if desc {
			queryBuilder.WriteString(" AND p.id < $3")
		} else {
			queryBuilder.WriteString(" AND p.id > $3")
		}
	}

	if desc {
		queryBuilder.WriteString(" ORDER BY p.created DESC, p.id DESC LIMIT $2")
	} else {
		queryBuilder.WriteString(" ORDER BY p.created, p.id LIMIT $2")
	}

	return r.byId(ctx, queryBuilder.String(), id, limit, since)
}

func (r *PostRepository) TreeByThreadSlug(ctx context.Context, slug string, limit int32, desc bool, since int64) ([]domain.Post, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(
		`	SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id
    			FROM posts p JOIN threads t ON t.id = p.thread_id`,
	)

	if since != 0 {
		if desc {
			queryBuilder.WriteString(" JOIN posts ON posts.id = $3 WHERE p.path || p.id < posts.path || posts.id")
		} else {
			queryBuilder.WriteString(" JOIN posts ON posts.id = $3 WHERE p.path || p.id > posts.path || posts.id")
		}
		queryBuilder.WriteString(" AND t.slug = $1 ORDER BY p.path || p.id")
	} else {
		queryBuilder.WriteString(" WHERE t.slug = $1 ORDER BY p.path || p.id")
	}

	if desc {
		queryBuilder.WriteString(" DESC")
	}
	queryBuilder.WriteString(" LIMIT $2")

	return r.bySlug(ctx, queryBuilder.String(), slug, limit, since)
}

func (r *PostRepository) TreeByThreadId(ctx context.Context, id int, limit int32, desc bool, since int64) ([]domain.Post, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(
		`	SELECT p.user_nn, p.created, (SELECT forum_slug FROM threads WHERE id = $1), p.id, p.message, p.parent_id, p.thread_id
    			FROM posts p`,
	)

	if since != 0 {
		if desc {
			queryBuilder.WriteString(" JOIN posts ON posts.id = $3 WHERE p.path || p.id < posts.path || posts.id")
		} else {
			queryBuilder.WriteString(" JOIN posts ON posts.id = $3 WHERE p.path || p.id > posts.path || posts.id")
		}
		queryBuilder.WriteString(" AND p.thread_id = $1 ORDER BY p.path || p.id")
	} else {
		queryBuilder.WriteString(" WHERE p.thread_id = $1 ORDER BY p.path || p.id")
	}

	if desc {
		queryBuilder.WriteString(" DESC")
	}
	queryBuilder.WriteString(" LIMIT $2")

	return r.byId(ctx, queryBuilder.String(), id, limit, since)
}

func (r *PostRepository) ParentTreeByThreadSlug(ctx context.Context, slug string, limit int32, desc bool, since int64) ([]domain.Post, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("WITH ranked_posts AS (SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id,p.path || p.id AS path,")

	if desc {
		queryBuilder.WriteString(" dense_rank() over (ORDER BY COALESCE(path [1], p.id) desc) AS rank")
	} else {
		queryBuilder.WriteString(" dense_rank() over (ORDER BY COALESCE(path [1], p.id)) AS rank")
	}
	queryBuilder.WriteString(
		`	FROM posts p JOIN threads t on p.thread_id = t.id WHERE t.slug = $1)
				SELECT p.user_nn, p.created, p.forum_slug, p.id, p.message, p.parent_id, p.thread_id 
				FROM ranked_posts p`)

	if since != 0 {
		queryBuilder.WriteString(
			`	JOIN ranked_posts posts ON posts.id = $3 
				WHERE p.rank <= $2 + posts.rank AND (p.rank > posts.rank OR p.rank = posts.rank AND p.path > posts.path) 
				ORDER BY p.rank, p.path`)
	} else {
		queryBuilder.WriteString(" WHERE p.rank <= $2 ORDER BY p.rank, p.path")
	}

	return r.bySlug(ctx, queryBuilder.String(), slug, limit, since)
}

func (r *PostRepository) ParentTreeByThreadId(ctx context.Context, id int, limit int32, desc bool, since int64) ([]domain.Post, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString("WITH ranked_posts AS (SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id,p.path || p.id AS path,")

	if desc {
		queryBuilder.WriteString(" dense_rank() over (ORDER BY COALESCE(path [1], p.id) desc) AS rank")
	} else {
		queryBuilder.WriteString(" dense_rank() over (ORDER BY COALESCE(path [1], p.id)) AS rank")
	}
	queryBuilder.WriteString(
		`	FROM posts p JOIN threads t on p.thread_id = t.id WHERE t.id = $1)
				SELECT p.user_nn, p.created, p.forum_slug, p.id, p.message, p.parent_id, p.thread_id 
				FROM ranked_posts p`)

	if since != 0 {
		queryBuilder.WriteString(
			`	JOIN ranked_posts posts ON posts.id = $3 
				WHERE p.rank <= $2 + posts.rank AND (p.rank > posts.rank OR p.rank = posts.rank AND p.path > posts.path) 
				ORDER BY p.rank, p.path`)
	} else {
		queryBuilder.WriteString(" WHERE p.rank <= $2 ORDER BY p.rank, p.path")
	}

	return r.byId(ctx, queryBuilder.String(), id, limit, since)
}

func (r *PostRepository) byId(ctx context.Context, query string, id int, limit int32, since int64) ([]domain.Post, error) {
	var rows pgx.Rows
	var err error
	if since != 0 {
		rows, err = r.DB.Query(ctx, query, id, limit, since)
	} else {
		rows, err = r.DB.Query(ctx, query, id, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("get post by id: %w", err)
	}
	defer rows.Close()

	posts := make([]domain.Post, 0, 1)
	for rows.Next() {
		var post domain.Post
		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			return nil, fmt.Errorf("scan posts: %w", err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan posts: %w", err)
	}
	rows.Close()

	if len(posts) == 0 {
		err := r.DB.QueryRow(ctx, "SELECT id FROM threads WHERE id = $1", id).Scan(&id)
		if err != nil {
			return nil, domain.ErrPostNotFoundThread
		}
	}

	return posts, nil
}

func (r *PostRepository) bySlug(ctx context.Context, query string, slug string, limit int32, since int64) ([]domain.Post, error) {
	var rows pgx.Rows
	var err error
	if since != 0 {
		rows, err = r.DB.Query(ctx, query, slug, limit, since)
	} else {
		rows, err = r.DB.Query(ctx, query, slug, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("get post by slug: %w", err)
	}
	defer rows.Close()

	posts := make([]domain.Post, 0, 1)
	for rows.Next() {
		var post domain.Post
		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			return nil, fmt.Errorf("get post by slug: %w", err)
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan posts: %w", err)
	}
	rows.Close()

	if len(posts) == 0 {
		err := r.DB.QueryRow(ctx, "SELECT slug FROM threads WHERE slug = $1", slug).Scan(&slug)
		if err != nil {
			return nil, domain.ErrPostNotFoundThread
		}
	}

	return posts, nil
}
