package handlers

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/viewsharp/technopark-forum/internal/resources/forum"
	"github.com/viewsharp/technopark-forum/internal/resources/post"
	"github.com/viewsharp/technopark-forum/internal/resources/thread"
	"github.com/viewsharp/technopark-forum/internal/resources/user"
	"github.com/viewsharp/technopark-forum/internal/resources/vote"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type StorageBundle struct {
	forum  *forum.Storage
	post   *post.Storage
	thread *thread.Storage
	user   *user.Storage
	vote   *vote.Storage
}

func NewStorageBundle(db DB) *StorageBundle {
	return &StorageBundle{
		forum:  &forum.Storage{DB: db},
		post:   &post.Storage{DB: db},
		thread: &thread.Storage{DB: db},
		user:   &user.Storage{DB: db},
		vote:   &vote.Storage{DB: db},
	}
}

func (sb *StorageBundle) DB() DB {
	return sb.forum.DB
}
