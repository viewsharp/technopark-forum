// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: forum.sql

package db

import (
	"context"
)

const createForum = `-- name: CreateForum :one
INSERT INTO forums (slug, title, user_nn)
VALUES ($1, $2, $3)
RETURNING slug, title, user_nn, posts, threads
`

type CreateForumParams struct {
	Slug   string
	Title  string
	UserNn string
}

func (q *Queries) CreateForum(ctx context.Context, arg CreateForumParams) (Forum, error) {
	row := q.db.QueryRow(ctx, createForum, arg.Slug, arg.Title, arg.UserNn)
	var i Forum
	err := row.Scan(
		&i.Slug,
		&i.Title,
		&i.UserNn,
		&i.Posts,
		&i.Threads,
	)
	return i, err
}

const getForumBySlug = `-- name: GetForumBySlug :one
SELECT slug, title, user_nn, posts, threads
FROM forums
WHERE slug = $1
`

func (q *Queries) GetForumBySlug(ctx context.Context, slug string) (Forum, error) {
	row := q.db.QueryRow(ctx, getForumBySlug, slug)
	var i Forum
	err := row.Scan(
		&i.Slug,
		&i.Title,
		&i.UserNn,
		&i.Posts,
		&i.Threads,
	)
	return i, err
}

const increasePostsCount = `-- name: IncreasePostsCount :exec
UPDATE forums
SET posts = posts + $1::INT
WHERE slug = $2
`

type IncreasePostsCountParams struct {
	NewPostsCount int32
	Slug          string
}

func (q *Queries) IncreasePostsCount(ctx context.Context, arg IncreasePostsCountParams) error {
	_, err := q.db.Exec(ctx, increasePostsCount, arg.NewPostsCount, arg.Slug)
	return err
}
