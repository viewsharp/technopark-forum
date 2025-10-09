package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/viewsharp/technopark-forum/internal/domain"
)

type ThreadRepository struct {
	DB Database
}

func (r *ThreadRepository) Add(ctx context.Context, thread *domain.Thread) error {
	err := r.DB.QueryRow(
		ctx,
		`	INSERT INTO threads (slug, created, title, message, user_nn, forum_slug)
            	VALUES ($1, $2, $3, $4, $5, (SELECT slug FROM forums WHERE slug = $6))
              	RETURNING id, forum_slug, slug`,
		thread.Slug, thread.Created, thread.Title, thread.Message, thread.Author, thread.Forum,
	).Scan(&thread.Id, &thread.Forum, &thread.Slug)

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
	return nil
}

func (r *ThreadRepository) BySlug(ctx context.Context, slug string) (*domain.Thread, error) {
	var result domain.Thread

	err := r.DB.QueryRow(
		ctx,
		`	SELECT id, slug, created, title, message, user_nn, forum_slug, votes
            	FROM threads
              	WHERE slug = $1`,
		slug,
	).Scan(&result.Id, &result.Slug, &result.Created, &result.Title, &result.Message, &result.Author, &result.Forum, &result.Votes)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get thread: %w", err)
	}
	return &result, nil
}

func (r *ThreadRepository) ById(ctx context.Context, id int) (*domain.Thread, error) {
	var result domain.Thread

	err := r.DB.QueryRow(
		ctx,
		`	SELECT id, slug, created, title, message, user_nn, forum_slug, votes
            	FROM threads
              	WHERE id = $1`,
		id,
	).Scan(&result.Id, &result.Slug, &result.Created, &result.Title, &result.Message, &result.Author, &result.Forum, &result.Votes)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get thread: %w", err)
	}
	return &result, nil
}

func (r *ThreadRepository) ByForumSlug(ctx context.Context, slug string, desc bool, since string, limit int32) (*domain.Threads, error) {
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

	result := make(domain.Threads, 0, limit)
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

		result = append(result, &thread)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan threads: %w", err)
	}
	rows.Close()

	if len(result) == 0 {
		var forumSlug *string
		err = r.DB.QueryRow(ctx, "SELECT slug FROM forums WHERE slug = $1", slug).Scan(&forumSlug)
		if forumSlug == nil {
			return nil, domain.ErrThreadNotFoundForum
		}
	}

	return &result, nil
}

func (r *ThreadRepository) UpdateById(ctx context.Context, id int, thread *domain.ThreadUpdate) error {
	_, err := r.DB.Exec(
		ctx,
		`	UPDATE threads 
				SET title = COALESCE($1, title), message = COALESCE($2, message)
				WHERE id = $3
				RETURNING title, message`,
		thread.Title, thread.Message, id,
	)

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
		return fmt.Errorf("select forum by slug: %w", err)
	}
	return nil
}

func (r *ThreadRepository) UpdateBySlug(ctx context.Context, slug string, thread *domain.ThreadUpdate) error {
	_, err := r.DB.Exec(
		ctx,
		`	UPDATE threads 
				SET title = COALESCE($1, title), message = COALESCE($2, message)
				WHERE slug = $3
				RETURNING title, message`,
		thread.Title, thread.Message, slug,
	)

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
		return fmt.Errorf("select forum by slug: %w", err)
	}
	return nil
}
