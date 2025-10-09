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
		Thread: &ThreadRepository{DB: database, Queries: queries},
		User:   &UserRepository{DB: database, Queries: queries},
		Vote:   &VoteRepository{DB: database, Queries: queries},
	}
}

func (rs *RepositorySet) DB() Database {
	return rs.Forum.DB
}
