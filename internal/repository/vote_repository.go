package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/viewsharp/technopark-forum/internal/db"
	"github.com/viewsharp/technopark-forum/internal/domain"
)

type VoteRepository struct {
	DB      Database
	Queries *db.Queries
}

func (r *VoteRepository) AddByThreadId(ctx context.Context, vote *domain.Vote, threadId int) error {
	err := r.Queries.UpsertVoteById(ctx, db.UpsertVoteByIdParams{
		ThreadID: int32(threadId),
		UserNn:   *vote.Nickname,
		Voice:    pgtype.Int4{Int32: *vote.Voice, Valid: true},
	})

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
	err := r.Queries.UpsertVoteBySlug(ctx, db.UpsertVoteBySlugParams{
		Slug:   pgtype.Text{String: threadSlug, Valid: true},
		UserNn: *vote.Nickname,
		Voice:  pgtype.Int4{Int32: *vote.Voice, Valid: true},
	})

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
