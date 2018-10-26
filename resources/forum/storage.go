package forum

import (
	"database/sql"
	"github.com/lib/pq"
)

type Storage struct {
	DB *sql.DB
}

func (s *Storage) Add(forum *Forum) error {
	err := s.DB.QueryRow(
		`	INSERT INTO forums (slug, title, user_nn)
            	VALUES ($1, $2, (SELECT nickname FROM users WHERE nickname=$3))
            	RETURNING user_nn`,
		forum.Slug, forum.Title, forum.User,
	).Scan(&forum.User)

	if err == nil {
		return nil
	}

	switch err.(*pq.Error).Code.Name() {
	case "unique_violation":
		return ErrUniqueViolation
	case "not_null_violation":
		return ErrNotFoundUser
	}

	return ErrUnknown
}

func (s *Storage) BySlug(slug string) (*Forum, error) {
	var result Forum

	err := s.DB.QueryRow(
		`	SELECT slug, title, user_nn
            	FROM forums
              	WHERE slug = $1`,
		slug,
	).Scan(&result.Slug, &result.Title, &result.User)

	if err == nil {
		return &result, nil
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return nil, ErrNotFound
	}

	return nil, ErrUnknown
}

func (s *Storage) FullBySlug(slug string) (*Forum, error) {
	var result Forum

	err := s.DB.QueryRow(
		`	SELECT 
					posts,
                	slug,
                	threads,
                	title,
                	user_nn
            	FROM forums
            	WHERE slug = $1`,
		slug,
	).Scan(&result.Posts, &result.Slug, &result.Threads, &result.Title, &result.User)

	if err == nil {
		return &result, nil
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return nil, ErrNotFound
	}

	return nil, ErrUnknown
}
