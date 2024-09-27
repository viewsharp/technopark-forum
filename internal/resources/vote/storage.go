package vote

import (
	"database/sql"

	"github.com/lib/pq"
)

type DB interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
}

type Storage struct {
	DB DB
}

func (s *Storage) AddByThreadId(vote *Vote, threadId int) error {
	_, err := s.DB.Exec(
		`
			INSERT INTO votes (thread_id, user_nn, voice) 
			VALUES ($1, $2, $3) 
			ON CONFLICT ON CONSTRAINT votes_thread_user_unique 
			DO UPDATE SET voice = $3
				WHERE votes.thread_id = (SELECT id FROM threads WHERE id = $1) AND votes.user_nn = $2;`,
		threadId, vote.Nickname, vote.Voice,
	)

	if err == nil {
		return nil
	}

	switch err.(*pq.Error).Code.Name() {
	case "foreign_key_violation":
		return ErrNotFoundUser
	case "not_null_violation":
		return ErrNotFoundThread
	}

	return ErrUnknown
}

func (s *Storage) AddByThreadSlug(vote *Vote, threadSlug string) error {
	_, err := s.DB.Exec(
		`
			INSERT INTO votes (thread_id, user_nn, voice) 
			VALUES ((SELECT id FROM threads WHERE slug = $1), $2, $3) 
			ON CONFLICT ON CONSTRAINT votes_thread_user_unique 
			DO UPDATE SET voice = $3
				WHERE votes.thread_id = (SELECT id FROM threads WHERE slug = $1) AND votes.user_nn = $2;`,
		threadSlug, vote.Nickname, vote.Voice,
	)

	if err == nil {
		return nil
	}

	switch err.(*pq.Error).Code.Name() {
	case "foreign_key_violation":
		return ErrNotFoundUser
	case "not_null_violation":
		return ErrNotFoundThread
	}

	return ErrUnknown
}

func (s *Storage) Sum() {

}
