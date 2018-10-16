package handlers

import (
	"database/sql"
	"github.com/viewsharp/TexPark_DBMSs/resources/forum"
	"github.com/viewsharp/TexPark_DBMSs/resources/post"
	"github.com/viewsharp/TexPark_DBMSs/resources/thread"
	"github.com/viewsharp/TexPark_DBMSs/resources/user"
	"github.com/viewsharp/TexPark_DBMSs/resources/vote"
)

type StorageBundle struct {
	forum *forum.Storage
	post *post.Storage
	thread *thread.Storage
	user *user.Storage
	vote *vote.Storage
}

func NewStorageBundle(db *sql.DB) *StorageBundle {
	return &StorageBundle{
		forum: &forum.Storage{DB:db},
		post: &post.Storage{DB:db},
		thread: &thread.Storage{DB:db},
		user: &user.Storage{DB:db},
		vote: &vote.Storage{DB:db},
	}
}

func (sb *StorageBundle)DB() *sql.DB {
	return sb.forum.DB
}