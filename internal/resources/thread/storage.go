package thread

import (
	"database/sql"
	"strings"

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

func (s *Storage) Add(thread *Thread) error {
	err := s.DB.QueryRow(
		`	INSERT INTO threads (slug, created, title, message, user_nn, forum_slug)
            	VALUES ($1, $2, $3, $4, $5, (SELECT slug FROM forums WHERE slug = $6))
              	RETURNING id, forum_slug, slug`,
		thread.Slug, thread.Created, thread.Title, thread.Message, thread.Author, thread.Forum,
	).Scan(&thread.Id, &thread.Forum, &thread.Slug)

	if err == nil {
		return nil
	}

	switch err.(*pq.Error).Code.Name() {
	case "unique_violation":
		return ErrUniqueViolation
	case "foreign_key_violation":
		return ErrNotFoundUser
	case "not_null_violation":
		return ErrNotFoundForum
	}

	return ErrUnknown
}

func (s *Storage) BySlug(slug string) (*Thread, error) {
	var result Thread

	err := s.DB.QueryRow(
		`	SELECT id, slug, created, title, message, user_nn, forum_slug, votes
            	FROM threads
              	WHERE slug = $1`,
		slug,
	).Scan(&result.Id, &result.Slug, &result.Created, &result.Title, &result.Message, &result.Author, &result.Forum, &result.Votes)

	if err == nil {
		return &result, nil
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return nil, ErrNotFound
	}

	return nil, ErrUnknown
}

func (s *Storage) ById(id int) (*Thread, error) {
	var result Thread

	err := s.DB.QueryRow(
		`	SELECT id, slug, created, title, message, user_nn, forum_slug, votes
            	FROM threads
              	WHERE id = $1`,
		id,
	).Scan(&result.Id, &result.Slug, &result.Created, &result.Title, &result.Message, &result.Author, &result.Forum, &result.Votes)

	if err == nil {
		return &result, nil
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return nil, ErrNotFound
	}

	return nil, ErrUnknown
}

func (s *Storage) ByForumSlug(slug string, desc bool, since string, limit int) (*Threads, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`	SELECT id, slug, created, title, message, user_nn, forum_slug, votes
            						FROM threads t
									WHERE forum_slug = $1`)

	if since != "" {
		if desc {
			queryBuilder.WriteString(" AND created <= $3")
		} else {
			queryBuilder.WriteString(" AND created >= $3")
		}
	}

	queryBuilder.WriteString(" ORDER BY created")
	if desc {
		queryBuilder.WriteString(" DESC")
	}

	queryBuilder.WriteString(" LIMIT $2")

	var rows *sql.Rows
	var err error
	if since == "" {
		rows, err = s.DB.Query(queryBuilder.String(), slug, limit)
	} else {
		rows, err = s.DB.Query(queryBuilder.String(), slug, limit, since)
	}
	defer rows.Close()

	if err != nil {
		return nil, ErrUnknown
	}

	result := make(Threads, 0, 1)
	for rows.Next() {
		var thread Thread
		err = rows.Scan(
			&thread.Id,
			&thread.Slug,
			&thread.Created,
			&thread.Title,
			&thread.Message,
			&thread.Author,
			&thread.Forum,
			&thread.Votes,
		)

		if err != nil {
			return nil, ErrUnknown
		}

		result = append(result, &thread)
	}

	if len(result) == 0 {
		var forumSlug *string
		err = s.DB.QueryRow("SELECT slug FROM forums WHERE slug = $1", slug).Scan(&forumSlug)
		if forumSlug == nil {
			return nil, ErrNotFoundForum
		}
	}

	return &result, nil
}

func (s *Storage) UpdateById(id int, thread *ThreadUpdate) error {
	_, err := s.DB.Exec(
		`	UPDATE threads 
				SET title = COALESCE($1, title), message = COALESCE($2, message)
				WHERE id = $3
				RETURNING title, message`,
		thread.Title, thread.Message, id,
	)

	if err == nil {
		return nil
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return ErrNotFound
	}

	switch err.(*pq.Error).Code.Name() {
	case "unique_violation":
		return ErrUniqueViolation
	}

	return ErrUnknown
}

func (s *Storage) UpdateBySlug(slug string, thread *ThreadUpdate) error {
	_, err := s.DB.Exec(
		`	UPDATE threads 
				SET title = COALESCE($1, title), message = COALESCE($2, message)
				WHERE slug = $3
				RETURNING title, message`,
		thread.Title, thread.Message, slug,
	)

	if err == nil {
		return nil
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return ErrNotFound
	}

	switch err.(*pq.Error).Code.Name() {
	case "unique_violation":
		return ErrUniqueViolation
	}

	return ErrUnknown
}
