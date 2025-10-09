-- name: GetThreadBySlug :one
SELECT *
FROM threads
WHERE slug = $1;

-- name: GetThreadByID :one
SELECT *
FROM threads
WHERE id = $1;

-- name: CreateThread :one
INSERT INTO threads (slug, created, title, message, user_nn, forum_slug)
VALUES ($1, $2, $3, $4, $5, (SELECT f.slug FROM forums f WHERE f.slug = $6))
RETURNING id, forum_slug, slug;

-- name: UpdateThreadById :exec
UPDATE threads 
SET title = COALESCE(sqlc.narg(title), title), 
    message = COALESCE(sqlc.narg(message), message)
WHERE id = $1;

-- name: UpdateThreadBySlug :exec
UPDATE threads 
SET title = COALESCE(sqlc.narg(title), title), 
    message = COALESCE(sqlc.narg(message), message)
WHERE slug = $1;

-- name: CheckThreadExistsById :one
SELECT id FROM threads WHERE id = $1;

-- name: CheckThreadExistsBySlug :one
SELECT slug FROM threads WHERE slug = $1;
