package post

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
	"github.com/viewsharp/technopark-forum/internal/usecase/forum"
	"github.com/viewsharp/technopark-forum/internal/usecase/thread"
	"github.com/viewsharp/technopark-forum/internal/usecase/user"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type Usecase struct {
	DB      DB
	Queries *db.Queries
}

var regexInvalidAuthor, _ = regexp.Compile(`^Key \(user_nn\)=\(([\w\.]+)\) is not present in table "users"\.$`)

func (s *Usecase) AddByThreadSlug(ctx context.Context, posts []Post, slug string) error {
	dbThread, err := s.Queries.GetThreadBySlug(ctx, pgtype.Text{String: slug, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFoundThread
		}
		return fmt.Errorf("select thread: %w", err)
	}

	return s.add(ctx, posts, dbThread.ID, dbThread.ForumSlug)
}

func (s *Usecase) AddByThreadId(ctx context.Context, posts []Post, threadId int32) error {
	dbThread, err := s.Queries.GetThreadByID(ctx, threadId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFoundThread
		}
		return fmt.Errorf("select thread: %w", err)
	}

	return s.add(ctx, posts, threadId, dbThread.ForumSlug)
}

func (s *Usecase) add(ctx context.Context, posts []Post, threadId int32, forumSlug string) error {
	// select parents

	parentIDMap := make(map[int32]struct{})
	for _, post := range posts {
		if post.Parent != nil {
			parentIDMap[*post.Parent] = struct{}{}
		}
	}

	parentIDs := slices.AppendSeq(make([]int32, 0, len(parentIDMap)), maps.Keys(parentIDMap))
	parents, err := s.Queries.ListByID(ctx, parentIDs)
	if err != nil {
		return fmt.Errorf("list parents by id: %w", err)
	}

	parentByID := make(map[int32]db.Post, len(parentIDs))
	for _, parent := range parents {
		parentByID[parent.ID] = parent
	}

	// insert posts

	postsParams := make([]db.CreatePostsParams, 0, len(posts))
	for _, post := range posts {
		var parentID pgtype.Int4
		var path []int32
		if post.Parent != nil {
			if parent, ok := parentByID[*post.Parent]; ok {
				if parent.ThreadID != threadId {
					return ErrInvalidParent
				}

				parentID = pgtype.Int4{Int32: parent.ID, Valid: true}
				path = append(parent.Path, parent.ID)
			} else {
				return ErrInvalidParent
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

	postsBatch := s.Queries.CreatePosts(ctx, postsParams)
	postsBatch.QueryRow(func(i int, post db.Post, batchErr error) {
		if errors.Is(batchErr, db.ErrBatchAlreadyClosed) {
			return
		}
		if batchErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(batchErr, &pgErr) && pgErr.Code == "23503" {
				err = ErrNotFoundUser{Nickname: post.UserNn}
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
		posts[i].Parent = &post.ParentID.Int32
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

	forumUsersBatch := s.Queries.CreateForumUser(ctx, forumUsersParams)
	forumUsersBatch.Exec(func(i int, batchErr error) {
		if batchErr != nil && !errors.Is(batchErr, db.ErrBatchAlreadyClosed) {
			err = batchErr
		}
		forumUsersBatch.Close()
	})

	// update posts count

	err = s.Queries.IncreasePostsCount(ctx, db.IncreasePostsCountParams{
		NewPostsCount: int32(len(posts)),
		Slug:          forumSlug,
	})
	if err != nil {
		return fmt.Errorf("update forum: %w", err)
	}

	return nil
}

func (s *Usecase) ById(ctx context.Context, id int, related []string) (*PostFull, error) {
	userObj := user.User{}
	forumObj := forum.Forum{}
	postObj := Post{}
	threadObj := thread.Thread{}
	result := PostFull{}

	err := s.DB.QueryRow(
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
			return nil, ErrNotFound
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

func (s *Usecase) UpdateById(ctx context.Context, id int, post PostUpdate) error {
	if post.Message == nil {
		return nil
	}

	_, err := s.DB.Exec(
		ctx,
		`	UPDATE posts 
				SET message = $1, isedited = TRUE
				WHERE id = $2`,
		post.Message, id,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("get post: %w", err)
	}

	return nil
}

func (s *Usecase) FlatByThreadSlug(ctx context.Context, slug string, limit int, desc bool, since int) ([]Post, error) {
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

	return s.bySlug(ctx, queryBuilder.String(), slug, limit, since)
}

func (s *Usecase) FlatByThreadId(ctx context.Context, id int, limit int, desc bool, since int) ([]Post, error) {
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

	return s.byId(ctx, queryBuilder.String(), id, limit, since)
}

func (s *Usecase) TreeByThreadSlug(ctx context.Context, slug string, limit int, desc bool, since int) ([]Post, error) {
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

	return s.bySlug(ctx, queryBuilder.String(), slug, limit, since)
}

func (s *Usecase) TreeByThreadId(ctx context.Context, id int, limit int, desc bool, since int) ([]Post, error) {
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

	return s.byId(ctx, queryBuilder.String(), id, limit, since)
}
func (s *Usecase) ParentTreeByThreadSlug(ctx context.Context, slug string, limit int, desc bool, since int) ([]Post, error) {
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

	return s.bySlug(ctx, queryBuilder.String(), slug, limit, since)
}

func (s *Usecase) ParentTreeByThreadId(ctx context.Context, id int, limit int, desc bool, since int) ([]Post, error) {
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

	return s.byId(ctx, queryBuilder.String(), id, limit, since)
}

func (s *Usecase) byId(ctx context.Context, query string, id int, limit int, since int) ([]Post, error) {
	var rows pgx.Rows
	var err error
	if since != 0 {
		rows, err = s.DB.Query(ctx, query, id, limit, since)
	} else {
		rows, err = s.DB.Query(ctx, query, id, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("get post by id: %w", err)
	}
	defer rows.Close()

	posts := make([]Post, 0, 1)
	for rows.Next() {
		var post Post
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
		err := s.DB.QueryRow(ctx, "SELECT id FROM threads WHERE id = $1", id).Scan(&id)
		if err != nil {
			return nil, ErrNotFoundThread
		}
	}

	return posts, nil
}

func (s *Usecase) bySlug(ctx context.Context, query string, slug string, limit int, since int) ([]Post, error) {
	var rows pgx.Rows
	var err error
	if since != 0 {
		rows, err = s.DB.Query(ctx, query, slug, limit, since)
	} else {
		rows, err = s.DB.Query(ctx, query, slug, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("get post by slug: %w", err)
	}
	defer rows.Close()

	posts := make([]Post, 0, 1)
	for rows.Next() {
		var post Post
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
		err := s.DB.QueryRow(ctx, "SELECT slug FROM threads WHERE slug = $1", slug).Scan(&slug)
		if err != nil {
			return nil, ErrNotFoundThread
		}
	}

	return posts, nil
}
