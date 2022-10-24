package user

import (
	"database/sql"
	"github.com/lib/pq"
	"strings"
)

type Storage struct {
	DB *sql.DB
}

func (s *Storage) Add(user *User) error {
	_, err := s.DB.Exec(
		"INSERT INTO users (nickname, fullname, email, about)	VALUES ($1, $2, $3, $4)",
		user.Nickname, user.FullName, user.Email, user.About,
	)

	if err == nil {
		return nil
	}

	switch err.(*pq.Error).Code.Name() {
	case "unique_violation":
		return ErrUniqueViolation
	default:
		return ErrUnknown
	}
}

func (s *Storage) ByNickname(nickname string) (*User, error) {
	var result User

	err := s.DB.QueryRow(
		"SELECT nickname, fullname, email, about FROM users WHERE nickname = $1",
		nickname,
	).Scan(&result.Nickname, &result.FullName, &result.Email, &result.About)

	if err == nil {
		return &result, nil
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return nil, ErrNotFound
	}

	return nil, ErrUnknown
}

func (s *Storage) ByEmail(email string) (*User, error) {
	var result User

	err := s.DB.QueryRow("SELECT nickname, fullname, email, about FROM users WHERE email = $1", email,
	).Scan(&result.Nickname, &result.FullName, &result.Email, &result.About)

	if err == nil {
		return &result, nil
	}

	switch err.Error() {
	case "sql: no rows in result set":
		return nil, ErrNotFound
	}

	return nil, ErrUnknown
}

func (s *Storage) UpdateByNickname(nickname string, user *UserUpdate) error {
	err := s.DB.QueryRow(
		"UPDATE users " +
			"SET fullname = COALESCE($1, fullname), email = COALESCE($2, email), about = COALESCE($3, about) " +
			"WHERE nickname = $4 " +
			"RETURNING fullname, email, about",
		user.FullName, user.Email, user.About, nickname,
	).Scan(&user.FullName, &user.Email, &user.About)

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

func (s *Storage) ByForumSlug(slug string, desc bool, since string, limit int) (*Users, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(
		"SELECT u.nickname, u.fullname, u.email, u.about " +
			"FROM forum_user fu " +
			"JOIN users u ON fu.user_id = u.id " +
			"WHERE fu.forum_slug = $1")

	if since != "" {
		if desc {
			queryBuilder.WriteString(" AND nickname < $3")
		} else {
			queryBuilder.WriteString(" AND nickname > $3")
		}
	}

	queryBuilder.WriteString(" ORDER BY nickname")
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

	if err != nil {
		return nil, ErrUnknown
	}
	defer rows.Close()

	result := make(Users, 0, 1)
	for rows.Next() {
		var user User
		err = rows.Scan(&user.Nickname, &user.FullName, &user.Email, &user.About)

		if err != nil {
			return nil, ErrUnknown
		}

		result = append(result, &user)
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
