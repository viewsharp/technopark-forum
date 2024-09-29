package forum

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/viewsharp/technopark-forum/internal/db"
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

func (s *Usecase) Add(ctx context.Context, forum Forum) (*Forum, error) {
	user, err := s.Queries.GetUserByNickname(ctx, *forum.User)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFoundUser
		}
		return nil, fmt.Errorf("get user by nickname: %w", err)
	}

	dbForum, err := s.Queries.CreateForum(ctx, db.CreateForumParams{
		Slug:   *forum.Slug,
		Title:  *forum.Title,
		UserNn: user.Nickname,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUniqueViolation
		}
		return nil, fmt.Errorf("insert forum: %w", err)
	}

	return &Forum{
		Posts:   &dbForum.Posts.Int32,
		Slug:    &dbForum.Slug,
		Threads: &dbForum.Threads.Int32,
		Title:   &dbForum.Title,
		User:    &dbForum.UserNn,
	}, nil
}

func (s *Usecase) BySlug(ctx context.Context, slug string) (*Forum, error) {
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

func (s *Usecase) FullBySlug(ctx context.Context, slug string) (*Forum, error) {
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
