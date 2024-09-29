package handlers

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/viewsharp/technopark-forum/internal/db"
	"github.com/viewsharp/technopark-forum/internal/usecase/forum"
	"github.com/viewsharp/technopark-forum/internal/usecase/post"
	"github.com/viewsharp/technopark-forum/internal/usecase/thread"
	"github.com/viewsharp/technopark-forum/internal/usecase/user"
	"github.com/viewsharp/technopark-forum/internal/usecase/vote"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type UsecaseSet struct {
	forum  *forum.Usecase
	post   *post.Usecase
	thread *thread.Usecase
	user   *user.Usecase
	vote   *vote.Usecase
}

func NewUsecaseSet(db DB, queries *db.Queries) *UsecaseSet {
	return &UsecaseSet{
		forum:  &forum.Usecase{DB: db, Queries: queries},
		post:   &post.Usecase{DB: db, Queries: queries},
		thread: &thread.Usecase{DB: db},
		user:   &user.Usecase{DB: db},
		vote:   &vote.Usecase{DB: db},
	}
}

func (sb *UsecaseSet) DB() DB {
	return sb.forum.DB
}
