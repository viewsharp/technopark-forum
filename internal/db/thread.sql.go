// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: thread.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getThreadByID = `-- name: GetThreadByID :one
SELECT id, slug, created, title, message, votes, user_nn, forum_slug
FROM threads
WHERE id = $1
`

func (q *Queries) GetThreadByID(ctx context.Context, id int32) (Thread, error) {
	row := q.db.QueryRow(ctx, getThreadByID, id)
	var i Thread
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Created,
		&i.Title,
		&i.Message,
		&i.Votes,
		&i.UserNn,
		&i.ForumSlug,
	)
	return i, err
}

const getThreadBySlug = `-- name: GetThreadBySlug :one
SELECT id, slug, created, title, message, votes, user_nn, forum_slug
FROM threads
WHERE slug = $1
`

func (q *Queries) GetThreadBySlug(ctx context.Context, slug pgtype.Text) (Thread, error) {
	row := q.db.QueryRow(ctx, getThreadBySlug, slug)
	var i Thread
	err := row.Scan(
		&i.ID,
		&i.Slug,
		&i.Created,
		&i.Title,
		&i.Message,
		&i.Votes,
		&i.UserNn,
		&i.ForumSlug,
	)
	return i, err
}
