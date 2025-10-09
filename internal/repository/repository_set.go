package repository

import (
	"github.com/viewsharp/technopark-forum/internal/db"
)

type RepositorySet struct {
	Forum  *ForumRepository
	Post   *PostRepository
	Thread *ThreadRepository
	User   *UserRepository
	Vote   *VoteRepository
}

func NewRepositorySet(database Database, queries *db.Queries) *RepositorySet {
	return &RepositorySet{
		Forum:  &ForumRepository{DB: database, Queries: queries},
		Post:   &PostRepository{DB: database, Queries: queries},
		Thread: &ThreadRepository{DB: database},
		User:   &UserRepository{DB: database, Queries: queries},
		Vote:   &VoteRepository{DB: database},
	}
}

func (rs *RepositorySet) DB() Database {
	return rs.Forum.DB
}
