package vote

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

func (s *Storage) AddByThreadId(ctx context.Context, vote *Vote, threadId int) error {
	_, err := s.DB.Exec(
		ctx,
		`
			INSERT INTO votes (thread_id, user_nn, voice) 
			VALUES ($1, $2, $3) 
			ON CONFLICT ON CONSTRAINT votes_thread_user_unique 
			DO UPDATE SET voice = $3
				WHERE votes.thread_id = (SELECT id FROM threads WHERE id = $1) AND votes.user_nn = $2;`,
		threadId, vote.Nickname, vote.Voice,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23502":
				return ErrNotFoundUser
			case "23503":
				return ErrNotFoundThread
			}
		}
		return fmt.Errorf("insert vote: %w", err)
	}
	return nil
}

func (s *Storage) AddByThreadSlug(ctx context.Context, vote *Vote, threadSlug string) error {
	_, err := s.DB.Exec(
		ctx,
		`
			INSERT INTO votes (thread_id, user_nn, voice) 
			VALUES ((SELECT id FROM threads WHERE slug = $1), $2, $3) 
			ON CONFLICT ON CONSTRAINT votes_thread_user_unique 
			DO UPDATE SET voice = $3
				WHERE votes.thread_id = (SELECT id FROM threads WHERE slug = $1) AND votes.user_nn = $2;`,
		threadSlug, vote.Nickname, vote.Voice,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23502":
				return ErrNotFoundUser
			case "23503":
				return ErrNotFoundThread
			}
		}
		return fmt.Errorf("insert vote: %w", err)
	}
	return nil
}
