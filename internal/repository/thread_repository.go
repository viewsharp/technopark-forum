package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/viewsharp/technopark-forum/internal/db"
	"github.com/viewsharp/technopark-forum/internal/domain"
)

type ThreadRepository struct {
	DB      Database
	Queries *db.Queries
}

func (r *ThreadRepository) Add(ctx context.Context, thread *domain.Thread) error {
	var slug pgtype.Text
	if thread.Slug != nil {
		slug = pgtype.Text{String: *thread.Slug, Valid: true}
	}

	var created pgtype.Timestamptz
	if thread.Created != nil {
		created = pgtype.Timestamptz{Time: *thread.Created, Valid: true}
	}

	var message pgtype.Text
	if thread.Message != "" {
		message = pgtype.Text{String: thread.Message, Valid: true}
	}

	dbThread, err := r.Queries.CreateThread(ctx, db.CreateThreadParams{
		Slug:    slug,
		Created: created,
		Title:   thread.Title,
		Message: message,
		UserNn:  thread.Author,
		Slug_2:  *thread.Forum,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23502":
				return domain.ErrThreadNotFoundUser
			case "23503":
				return domain.ErrThreadNotFoundForum
			case "23505":
				return domain.ErrUniqueViolation
			}
		}
		return fmt.Errorf("insert threads: %w", err)
	}

	thread.Id = &dbThread.ID
	thread.Forum = &dbThread.ForumSlug
	thread.Slug = &dbThread.Slug.String

	return nil
}

func (r *ThreadRepository) BySlug(ctx context.Context, slug string) (*domain.Thread, error) {
	dbThread, err := r.Queries.GetThreadBySlug(ctx, pgtype.Text{String: slug, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get thread: %w", err)
	}

	slugStr := dbThread.Slug.String
	forumSlug := dbThread.ForumSlug
	votes := dbThread.Votes.Int32

	return &domain.Thread{
		Id:      &dbThread.ID,
		Slug:    &slugStr,
		Created: &dbThread.Created.Time,
		Title:   dbThread.Title,
		Message: dbThread.Message.String,
		Author:  dbThread.UserNn,
		Forum:   &forumSlug,
		Votes:   &votes,
	}, nil
}

func (r *ThreadRepository) ById(ctx context.Context, id int) (*domain.Thread, error) {
	dbThread, err := r.Queries.GetThreadByID(ctx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get thread: %w", err)
	}

	slugStr := dbThread.Slug.String
	forumSlug := dbThread.ForumSlug
	votes := dbThread.Votes.Int32

	return &domain.Thread{
		Id:      &dbThread.ID,
		Slug:    &slugStr,
		Created: &dbThread.Created.Time,
		Title:   dbThread.Title,
		Message: dbThread.Message.String,
		Author:  dbThread.UserNn,
		Forum:   &forumSlug,
		Votes:   &votes,
	}, nil
}

func (r *ThreadRepository) ByForumSlug(ctx context.Context, slug string, desc bool, since string, limit int32) ([]domain.Thread, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`	SELECT id, slug, created, title, message, user_nn, forum_slug, votes
            						FROM threads t
									WHERE forum_slug = $1`)

	if since != "" {
		if desc {
			queryBuilder.WriteString(" AND created <= $3")
		} else {
			queryBuilder.WriteString(" AND created >= $3")
		}
	}

	queryBuilder.WriteString(" ORDER BY created")
	if desc {
		queryBuilder.WriteString(" DESC")
	}

	queryBuilder.WriteString(" LIMIT $2")

	var rows pgx.Rows
	var err error
	if since == "" {
		rows, err = r.DB.Query(ctx, queryBuilder.String(), slug, limit)
	} else {
		rows, err = r.DB.Query(ctx, queryBuilder.String(), slug, limit, since)
	}
	if err != nil {
		return nil, fmt.Errorf("select thread: %w", err)
	}
	defer rows.Close()

	result := make([]domain.Thread, 0, limit)
	for rows.Next() {
		var thread domain.Thread
		err = rows.Scan(
			&thread.Id,
			&thread.Slug,
			&thread.Created,
			&thread.Title,
			&thread.Message,
			&thread.Author,
			&thread.Forum,
			&thread.Votes,
		)
		if err != nil {
			return nil, fmt.Errorf("scan thread: %w", err)
		}

		result = append(result, thread)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan threads: %w", err)
	}
	rows.Close()

	if len(result) == 0 {
		_, err = r.Queries.CheckForumExists(ctx, slug)
		if err != nil {
			return nil, domain.ErrThreadNotFoundForum
		}
	}

	return result, nil
}

func (r *ThreadRepository) UpdateById(ctx context.Context, id int, thread *domain.ThreadUpdate) error {
	var title pgtype.Text
	if thread.Title != nil {
		title = pgtype.Text{String: *thread.Title, Valid: true}
	}

	var message pgtype.Text
	if thread.Message != nil {
		message = pgtype.Text{String: *thread.Message, Valid: true}
	}

	err := r.Queries.UpdateThreadById(ctx, db.UpdateThreadByIdParams{
		ID:      int32(id),
		Title:   title,
		Message: message,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return domain.ErrUniqueViolation
			}
		}
		return fmt.Errorf("update thread by id: %w", err)
	}
	return nil
}

func (r *ThreadRepository) UpdateBySlug(ctx context.Context, slug string, thread *domain.ThreadUpdate) error {
	var title pgtype.Text
	if thread.Title != nil {
		title = pgtype.Text{String: *thread.Title, Valid: true}
	}

	var message pgtype.Text
	if thread.Message != nil {
		message = pgtype.Text{String: *thread.Message, Valid: true}
	}

	err := r.Queries.UpdateThreadBySlug(ctx, db.UpdateThreadBySlugParams{
		Slug:    pgtype.Text{String: slug, Valid: true},
		Title:   title,
		Message: message,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return domain.ErrUniqueViolation
			}
		}
		return fmt.Errorf("update thread by slug: %w", err)
	}
	return nil
}
