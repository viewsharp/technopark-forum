package vote

import (
	"database/sql"
)

type Storage struct {
	DB *sql.DB
}

func (s *Storage) AddByThreadId(vote *Vote, threadId int) error {
	_, err := s.DB.Exec(
		`
			INSERT INTO votes (thread_id, user_nn, voice) 
			VALUES ($1, $2, $3) 
			ON CONFLICT ON CONSTRAINT votes_thread_user_unique 
			DO UPDATE SET voice = $3
				WHERE votes.thread_id = $1 AND votes.user_nn = $2;`,
		threadId, vote.Nickname, vote.Voice,
	)

	if err == nil {
		return nil
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

	return ErrUnknown
}

func (s *Storage)Sum()  {

}
