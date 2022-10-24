package handlers

import (
	"database/sql"
	"github.com/viewsharp/technopark-forum/internal/resources/forum"
	"github.com/viewsharp/technopark-forum/internal/resources/post"
	"github.com/viewsharp/technopark-forum/internal/resources/thread"
	"github.com/viewsharp/technopark-forum/internal/resources/user"
	"github.com/viewsharp/technopark-forum/internal/resources/vote"
)

type StorageBundle struct {
	forum  *forum.Storage
	post   *post.Storage
	thread *thread.Storage
	user   *user.Storage
	vote   *vote.Storage
}

func NewStorageBundle(db *sql.DB) *StorageBundle {
	return &StorageBundle{
		forum:  &forum.Storage{DB: db},
		post:   &post.Storage{DB: db},
		thread: &thread.Storage{DB: db},
		user:   &user.Storage{DB: db},
		vote:   &vote.Storage{DB: db},
	}
}

func (sb *StorageBundle) DB() *sql.DB {
	return sb.forum.DB
}
