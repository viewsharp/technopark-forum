// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: user.sql

package db

import (
	"context"
)

const getUserByNickname = `-- name: GetUserByNickname :one
SELECT id, nickname, fullname, email, about FROM users WHERE nickname = $1
`

func (q *Queries) GetUserByNickname(ctx context.Context, nickname string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByNickname, nickname)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Nickname,
		&i.Fullname,
		&i.Email,
		&i.About,
	)
	return i, err
}