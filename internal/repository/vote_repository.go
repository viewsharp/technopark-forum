package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/viewsharp/technopark-forum/internal/domain"
)

type VoteRepository struct {
	DB Database
}

func (r *VoteRepository) AddByThreadId(ctx context.Context, vote *domain.Vote, threadId int) error {
	_, err := r.DB.Exec(
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
				return domain.ErrVoteNotFoundUser
			case "23503":
				return domain.ErrVoteNotFoundThread
			}
		}
		return fmt.Errorf("insert vote: %w", err)
	}
	return nil
}

func (r *VoteRepository) AddByThreadSlug(ctx context.Context, vote *domain.Vote, threadSlug string) error {
	_, err := r.DB.Exec(
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
				return domain.ErrVoteNotFoundUser
			case "23503":
				return domain.ErrVoteNotFoundThread
			}
		}
		return fmt.Errorf("insert vote: %w", err)
	}
	return nil
}
