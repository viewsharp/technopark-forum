package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/viewsharp/technopark-forum/internal/db"
	"github.com/viewsharp/technopark-forum/internal/domain"
)

type ForumRepository struct {
	DB      Database
	Queries *db.Queries
}

func (r *ForumRepository) Add(ctx context.Context, forum domain.Forum) (*domain.Forum, error) {
	user, err := r.Queries.GetUserByNickname(ctx, forum.User)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrForumNotFoundUser
		}
		return nil, fmt.Errorf("get user by nickname: %w", err)
	}

	dbForum, err := r.Queries.CreateForum(ctx, db.CreateForumParams{
		Slug:   forum.Slug,
		Title:  forum.Title,
		UserNn: user.Nickname,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrUniqueViolation
		}
		return nil, fmt.Errorf("insert forum: %w", err)
	}

	return &domain.Forum{
		Posts:   &dbForum.Posts.Int64,
		Slug:    dbForum.Slug,
		Threads: &dbForum.Threads.Int32,
		Title:   dbForum.Title,
		User:    dbForum.UserNn,
	}, nil
}

func (r *ForumRepository) BySlug(ctx context.Context, slug string) (*domain.Forum, error) {
	dbForum, err := r.Queries.GetForumBySlugLight(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select forum by slug: %w", err)
	}

	return &domain.Forum{
		Slug:  dbForum.Slug,
		Title: dbForum.Title,
		User:  dbForum.UserNn,
	}, nil
}

func (r *ForumRepository) FullBySlug(ctx context.Context, slug string) (*domain.Forum, error) {
	dbForum, err := r.Queries.GetForumBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select forum by slug: %w", err)
	}

	return &domain.Forum{
		Posts:   &dbForum.Posts.Int64,
		Slug:    dbForum.Slug,
		Threads: &dbForum.Threads.Int32,
		Title:   dbForum.Title,
		User:    dbForum.UserNn,
	}, nil
}
