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

type UserRepository struct {
	DB      Database
	Queries *db.Queries
}

func (r *UserRepository) Add(ctx context.Context, user *domain.User) error {
	var about pgtype.Text
	if user.About != nil {
		about = pgtype.Text{String: *user.About, Valid: true}
	}

	err := r.Queries.CreateUser(ctx, db.CreateUserParams{
		Nickname: user.Nickname,
		Fullname: user.FullName,
		Email:    user.Email,
		About:    about,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUniqueViolation
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *UserRepository) ByNickname(ctx context.Context, nickname string) (*domain.User, error) {
	dbUser, err := r.Queries.GetUserByNickname(ctx, nickname)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("select user: %w", err)
	}

	var about *string
	if dbUser.About.Valid {
		about = &dbUser.About.String
	}

	return &domain.User{
		Nickname: dbUser.Nickname,
		FullName: dbUser.Fullname,
		Email:    dbUser.Email,
		About:    about,
	}, nil
}

func (r *UserRepository) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	dbUser, err := r.Queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("select user: %w", err)
	}

	var about *string
	if dbUser.About.Valid {
		about = &dbUser.About.String
	}

	return &domain.User{
		Nickname: dbUser.Nickname,
		FullName: dbUser.Fullname,
		Email:    dbUser.Email,
		About:    about,
	}, nil
}

func (r *UserRepository) UpdateByNickname(ctx context.Context, nickname string, user *domain.UserUpdate) (*domain.User, error) {
	params := db.UpdateUserByNicknameParams{
		Nickname: nickname,
	}

	if user.FullName != nil {
		params.Fullname = pgtype.Text{String: *user.FullName, Valid: true}
	}
	if user.Email != nil {
		params.Email = pgtype.Text{String: *user.Email, Valid: true}
	}
	if user.About != nil {
		params.About = pgtype.Text{String: *user.About, Valid: true}
	}

	dbUser, err := r.Queries.UpdateUserByNickname(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrUniqueViolation
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}

	var about *string
	if dbUser.About.Valid {
		about = &dbUser.About.String
	}

	return &domain.User{
		Nickname: dbUser.Nickname,
		FullName: dbUser.Fullname,
		Email:    dbUser.Email,
		About:    about,
	}, nil
}

func (r *UserRepository) ByForumSlug(ctx context.Context, slug string, desc bool, since string, limit int32) (*domain.Users, error) {
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
		rows, err = r.DB.Query(ctx, queryBuilder.String(), slug, limit)
	} else {
		rows, err = r.DB.Query(ctx, queryBuilder.String(), slug, limit, since)
	}

	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}
	defer rows.Close()

	result := make(domain.Users, 0, 1)
	for rows.Next() {
		var user domain.User
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
		_, err = r.Queries.CheckForumExists(ctx, slug)
		if err != nil {
			return nil, domain.ErrUserNotFoundForum
		}
	}

	return &result, nil
}
