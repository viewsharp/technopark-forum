package user

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"strings"
)

type Storage struct {
	DB *sql.DB
}

func (s *Storage) Add(user *User) error {
	_, err := s.DB.Exec(
		`
			INSERT INTO users (nickname, fullname, email, about) 
			VALUES ($1, $2, $3, $4)`,
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
		` SELECT nickname, fullname, email, about 
				FROM users 
				WHERE nickname = $1 `,
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

	err := s.DB.QueryRow(
		`	SELECT nickname, fullname, email, about 
				FROM users 
				WHERE email = $1`,
		email,
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
		`	UPDATE users 
				SET fullname = COALESCE($1, fullname), email = COALESCE($2, email), about = COALESCE($3, about) 
				WHERE nickname = $4
				RETURNING fullname, email, about`,
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

func (s *Storage)ByForumSlug(slug string, desc bool, since string, limit int) (*Users, error) {
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`	SELECT *
									FROM (SELECT u.nickname, u.fullname, u.email, u.about
										FROM users u
									        JOIN threads t ON u.nickname = t.user_nn
									    WHERE t.forum_slug = $1
									    UNION
									    SELECT u.nickname, u.fullname, u.email, u.about
									    FROM users u
									        JOIN posts p ON u.nickname = p.user_nn
									    	JOIN threads t ON t.id = p.thread_id
									    WHERE t.forum_slug = $1) forum_users`)

	if since != "" {
		if desc {
			queryBuilder.WriteString(` WHERE nickname COLLATE "ucs_basic" < $3 COLLATE "ucs_basic"`)
		} else {
			queryBuilder.WriteString(` WHERE nickname COLLATE "ucs_basic" > $3 COLLATE "ucs_basic"`)
		}
	}

	queryBuilder.WriteString(` ORDER BY nickname COLLATE "ucs_basic"`)
	if desc {
		queryBuilder.WriteString(" DESC")
	}

	queryBuilder.WriteString(" LIMIT $2")

	fmt.Println(queryBuilder.String())

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