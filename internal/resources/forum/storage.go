package forum

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type Storage struct {
	DB DB
}

func (s *Storage) Add(ctx context.Context, forum *Forum) error {
	err := s.DB.QueryRow(
		ctx,
		`	INSERT INTO forums (slug, title, user_nn)
            	VALUES ($1, $2, (SELECT nickname FROM users WHERE nickname=$3))
            	RETURNING user_nn`,
		forum.Slug, forum.Title, forum.User,
	).Scan(&forum.User)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return ErrUniqueViolation
			case "23502":
				return ErrNotFoundUser
			}
		}
		return fmt.Errorf("insert forum: %w", err)
	}
	return nil
}

func (s *Storage) BySlug(ctx context.Context, slug string) (*Forum, error) {
	var result Forum

	err := s.DB.QueryRow(ctx, "SELECT slug, title, user_nn FROM forums WHERE slug = $1",
		slug,
	).Scan(&result.Slug, &result.Title, &result.User)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("select forum by slug: %w", err)
	}
	return &result, nil
}

func (s *Storage) FullBySlug(ctx context.Context, slug string) (*Forum, error) {
	var result Forum

	err := s.DB.QueryRow(
		ctx,
		"SELECT posts, slug, threads, title, user_nn FROM forums WHERE slug = $1",
		slug,
	).Scan(&result.Posts, &result.Slug, &result.Threads, &result.Title, &result.User)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("select forum by slug: %w", err)
	}
	return &result, nil
}
