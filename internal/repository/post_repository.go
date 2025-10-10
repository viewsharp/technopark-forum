package repository

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	sq "github.com/Masterminds/squirrel"
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
			Message:  post.Message,
			ParentID: parentID,
			UserNn:   post.Author,
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

		posts[i].Message = post.Message
		posts[i].Author = post.UserNn
		posts[i].Created = &post.Created.Time
		posts[i].Id = &post.ID
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
			Nickname:  post.Author,
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
	result := domain.PostFull{}

	needUser := slices.Contains(related, "user")
	needThread := slices.Contains(related, "thread")
	needForum := slices.Contains(related, "forum")

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Build base query
	query := psql.
		Select("p.user_nn", "p.created", "p.id", "p.isedited", "p.message", "p.parent_id", "p.thread_id").
		From("posts p").
		Where(sq.Eq{"p.id": id})

	// Add user fields if needed
	if needUser {
		query = query.
			Columns("u.about", "u.email", "u.fullname", "u.nickname").
			Join("users u ON p.user_nn = u.nickname")
	}

	// Add thread join if needed for thread or forum
	if needThread || needForum {
		query = query.Join("threads t ON p.thread_id = t.id")
	}

	// Add thread fields if needed
	if needThread {
		query = query.
			Columns("t.user_nn", "t.created", "t.id", "t.message", "t.slug", "t.title", "t.votes")
	}

	// Add forum fields if needed
	if needForum {
		query = query.
			Columns("f.posts", "f.slug", "f.threads", "f.title", "f.user_nn").
			Join("forums f ON t.forum_slug = f.slug")
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	// Prepare scan destinations
	var postObj domain.Post
	scanDest := []interface{}{
		&postObj.Author,
		&postObj.Created,
		&postObj.Id,
		&postObj.IsEdited,
		&postObj.Message,
		&postObj.Parent,
		&postObj.Thread,
	}

	var userObj domain.User
	if needUser {
		scanDest = append(scanDest,
			&userObj.About,
			&userObj.Email,
			&userObj.FullName,
			&userObj.Nickname,
		)
	}

	var threadObj domain.Thread
	if needThread {
		scanDest = append(scanDest,
			&threadObj.Author,
			&threadObj.Created,
			&threadObj.Id,
			&threadObj.Message,
			&threadObj.Slug,
			&threadObj.Title,
			&threadObj.Votes,
		)
	}

	var forumObj domain.Forum
	if needForum {
		scanDest = append(scanDest,
			&forumObj.Posts,
			&forumObj.Slug,
			&forumObj.Threads,
			&forumObj.Title,
			&forumObj.User,
		)
	}

	// Execute query
	err = r.DB.QueryRow(ctx, sql, args...).Scan(scanDest...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get post: %w", err)
	}

	// Get forum slug for post - we need thread info for this
	if postObj.Thread != nil {
		thread, err := r.Queries.GetThreadByID(ctx, *postObj.Thread)
		if err == nil {
			postObj.Forum = &thread.ForumSlug
		}
	}

	// Build result
	result.Post = &postObj

	if needUser {
		result.Author = &userObj
	}

	if needThread && threadObj.Id != nil {
		// Set forum slug for thread if not already set
		if threadObj.Forum == nil && postObj.Forum != nil {
			threadObj.Forum = postObj.Forum
		}
		result.Thread = &threadObj
	}

	if needForum {
		result.Forum = &forumObj
	}

	return &result, nil
}

func (r *PostRepository) UpdateById(ctx context.Context, id int64, post domain.PostUpdate) error {
	if post.Message == nil {
		return nil
	}

	err := r.Queries.UpdatePostMessage(ctx, db.UpdatePostMessageParams{
		Message: *post.Message,
		ID:      id,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("update post: %w", err)
	}

	return nil
}

func (r *PostRepository) FlatByThreadSlug(ctx context.Context, slug string, limit int32, desc bool, since int64) ([]domain.Post, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.
		Select("p.user_nn", "p.created", "t.forum_slug", "p.id", "p.message", "p.parent_id", "p.thread_id").
		From("posts p").
		Join("threads t ON p.thread_id = t.id").
		Where(sq.Eq{"t.slug": slug}).
		Limit(uint64(limit))

	if since != 0 {
		if desc {
			query = query.Where(sq.Lt{"p.id": since})
		} else {
			query = query.Where(sq.Gt{"p.id": since})
		}
	}

	if desc {
		query = query.OrderBy("p.created DESC", "p.id DESC")
	} else {
		query = query.OrderBy("p.created", "p.id")
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	return r.executePostQuery(ctx, sql, args, slug, 0)
}

func (r *PostRepository) FlatByThreadId(ctx context.Context, id int, limit int32, desc bool, since int64) ([]domain.Post, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.
		Select("p.user_nn", "p.created", "t.forum_slug", "p.id", "p.message", "p.parent_id", "p.thread_id").
		From("posts p").
		Join("threads t ON p.thread_id = t.id").
		Where(sq.Eq{"t.id": id}).
		Limit(uint64(limit))

	if since != 0 {
		if desc {
			query = query.Where(sq.Lt{"p.id": since})
		} else {
			query = query.Where(sq.Gt{"p.id": since})
		}
	}

	if desc {
		query = query.OrderBy("p.created DESC", "p.id DESC")
	} else {
		query = query.OrderBy("p.created", "p.id")
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	return r.executePostQuery(ctx, sql, args, "", id)
}

func (r *PostRepository) TreeByThreadSlug(ctx context.Context, slug string, limit int32, desc bool, since int64) ([]domain.Post, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.
		Select("p.user_nn", "p.created", "t.forum_slug", "p.id", "p.message", "p.parent_id", "p.thread_id").
		From("posts p").
		Join("threads t ON t.id = p.thread_id").
		Where(sq.Eq{"t.slug": slug}).
		Limit(uint64(limit))

	if since != 0 {
		query = query.Join("posts ON posts.id = ?", since)
		if desc {
			query = query.Where(sq.Expr("p.path || p.id < posts.path || posts.id"))
		} else {
			query = query.Where(sq.Expr("p.path || p.id > posts.path || posts.id"))
		}
	}

	if desc {
		query = query.OrderBy("p.path || p.id DESC")
	} else {
		query = query.OrderBy("p.path || p.id")
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	return r.executePostQuery(ctx, sql, args, slug, 0)
}

func (r *PostRepository) TreeByThreadId(ctx context.Context, id int, limit int32, desc bool, since int64) ([]domain.Post, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query := psql.
		Select("p.user_nn", "p.created").
		Column(sq.Expr("(SELECT forum_slug FROM threads WHERE id = ?) AS forum_slug", id)).
		Columns("p.id", "p.message", "p.parent_id", "p.thread_id").
		From("posts p").
		Where(sq.Eq{"p.thread_id": id}).
		Limit(uint64(limit))

	if since != 0 {
		query = query.Join("posts ON posts.id = ?", since)
		if desc {
			query = query.Where(sq.Expr("p.path || p.id < posts.path || posts.id"))
		} else {
			query = query.Where(sq.Expr("p.path || p.id > posts.path || posts.id"))
		}
	}

	if desc {
		query = query.OrderBy("p.path || p.id DESC")
	} else {
		query = query.OrderBy("p.path || p.id")
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	return r.executePostQuery(ctx, sql, args, "", id)
}

func (r *PostRepository) ParentTreeByThreadSlug(ctx context.Context, slug string, limit int32, desc bool, since int64) ([]domain.Post, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Build CTE rank expression
	var rankExpr string
	if desc {
		rankExpr = "dense_rank() over (ORDER BY COALESCE(path[1], p.id) desc) AS rank"
	} else {
		rankExpr = "dense_rank() over (ORDER BY COALESCE(path[1], p.id)) AS rank"
	}

	// Build main query
	query := psql.
		Select("p.user_nn", "p.created", "p.forum_slug", "p.id", "p.message", "p.parent_id", "p.thread_id").
		Prefix("WITH ranked_posts AS (SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id, p.path || p.id AS path, "+rankExpr+" FROM posts p JOIN threads t on p.thread_id = t.id WHERE t.slug = ?)", slug).
		From("ranked_posts p").
		OrderBy("p.rank", "p.path")

	if since != 0 {
		query = query.
			JoinClause("JOIN ranked_posts posts ON posts.id = ?", since).
			Where(sq.Expr("p.rank <= ? + posts.rank AND (p.rank > posts.rank OR p.rank = posts.rank AND p.path > posts.path)", limit))
	} else {
		query = query.
			Where(sq.LtOrEq{"p.rank": limit})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	return r.executePostQuery(ctx, sql, args, slug, 0)
}

func (r *PostRepository) ParentTreeByThreadId(ctx context.Context, id int, limit int32, desc bool, since int64) ([]domain.Post, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Build CTE rank expression
	var rankExpr string
	if desc {
		rankExpr = "dense_rank() over (ORDER BY COALESCE(path[1], p.id) desc) AS rank"
	} else {
		rankExpr = "dense_rank() over (ORDER BY COALESCE(path[1], p.id)) AS rank"
	}

	// Build main query
	query := psql.
		Select("p.user_nn", "p.created", "p.forum_slug", "p.id", "p.message", "p.parent_id", "p.thread_id").
		Prefix("WITH ranked_posts AS (SELECT p.user_nn, p.created, t.forum_slug, p.id, p.message, p.parent_id, p.thread_id, p.path || p.id AS path, "+rankExpr+" FROM posts p JOIN threads t on p.thread_id = t.id WHERE t.id = ?)", id).
		From("ranked_posts p").
		OrderBy("p.rank", "p.path")

	if since != 0 {
		query = query.
			JoinClause("JOIN ranked_posts posts ON posts.id = ?", since).
			Where(sq.Expr("p.rank <= ? + posts.rank AND (p.rank > posts.rank OR p.rank = posts.rank AND p.path > posts.path)", limit))
	} else {
		query = query.Where(sq.LtOrEq{"p.rank": limit})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	return r.executePostQuery(ctx, sql, args, "", id)
}

func (r *PostRepository) executePostQuery(ctx context.Context, query string, args []interface{}, slug string, threadId int) ([]domain.Post, error) {
	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
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

	if len(posts) == 0 {
		// Check if thread exists
		if slug != "" {
			_, err := r.Queries.CheckThreadExistsBySlug(ctx, pgtype.Text{String: slug, Valid: true})
			if err != nil {
				return nil, domain.ErrPostNotFoundThread
			}
		} else if threadId != 0 {
			_, err := r.Queries.CheckThreadExistsById(ctx, int32(threadId))
			if err != nil {
				return nil, domain.ErrPostNotFoundThread
			}
		}
	}

	return posts, nil
}
