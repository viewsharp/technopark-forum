package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

func (s *Storage) Add(ctx context.Context, user *User) error {
	_, err := s.DB.Exec(
		ctx,
		"INSERT INTO users (nickname, fullname, email, about)	VALUES ($1, $2, $3, $4)",
		user.Nickname, user.FullName, user.Email, user.About,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUniqueViolation
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (s *Storage) ByNickname(ctx context.Context, nickname string) (*User, error) {
	var result User

	err := s.DB.QueryRow(
		ctx,
		"SELECT nickname, fullname, email, about FROM users WHERE nickname = $1",
		nickname,
	).Scan(&result.Nickname, &result.FullName, &result.Email, &result.About)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("select user: %w", err)
	}

	return &result, nil
}

func (s *Storage) ByEmail(ctx context.Context, email string) (*User, error) {
	var result User

	err := s.DB.QueryRow(ctx, "SELECT nickname, fullname, email, about FROM users WHERE email = $1", email).Scan(&result.Nickname, &result.FullName, &result.Email, &result.About)
	if err != nil {
		return nil, fmt.Errorf("select user: %w", err)
	}

	return &result, nil
}

func (s *Storage) UpdateByNickname(ctx context.Context, nickname string, user *UserUpdate) error {
	err := s.DB.QueryRow(
		ctx,
		"UPDATE users "+
			"SET fullname = COALESCE($1, fullname), email = COALESCE($2, email), about = COALESCE($3, about) "+
			"WHERE nickname = $4 "+
			"RETURNING fullname, email, about",
		user.FullName, user.Email, user.About, nickname,
	).Scan(&user.FullName, &user.Email, &user.About)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUniqueViolation
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

func (s *Storage) ByForumSlug(ctx context.Context, slug string, desc bool, since string, limit int) (*Users, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(
		"SELECT u.nickname, u.fullname, u.email, u.about " +
			"FROM forum_user fu " +
			"JOIN users u ON fu.user_id = u.id " +
			"WHERE fu.forum_slug = $1")

	if since != "" {
		if desc {
			queryBuilder.WriteString(" AND nickname < $3")
		} else {
			queryBuilder.WriteString(" AND nickname > $3")
		}
	}

	queryBuilder.WriteString(" ORDER BY nickname")
	if desc {
		queryBuilder.WriteString(" DESC")
	}

	queryBuilder.WriteString(" LIMIT $2")

	var rows pgx.Rows
	var err error
	if since == "" {
		rows, err = s.DB.Query(ctx, queryBuilder.String(), slug, limit)
	} else {
		rows, err = s.DB.Query(ctx, queryBuilder.String(), slug, limit, since)
	}

	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	defer rows.Close()

	result := make(Users, 0, 1)
	for rows.Next() {
		var user User
		err = rows.Scan(&user.Nickname, &user.FullName, &user.Email, &user.About)
		if err != nil {
			return nil, fmt.Errorf("scan users %w", err)
		}

		result = append(result, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan users: %w", err)
	}
	rows.Close()

	if len(result) == 0 {
		var forumSlug *string
		err = s.DB.QueryRow(ctx, "SELECT slug FROM forums WHERE slug = $1", slug).Scan(&forumSlug)
		if forumSlug == nil {
			return nil, ErrNotFoundForum
		}
	}

	return &result, nil
}
