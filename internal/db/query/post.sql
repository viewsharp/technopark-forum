-- name: CreatePosts :batchone
INSERT INTO posts (message, parent_id, user_nn, thread_id, path)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListByID :many
SELECT *
FROM posts
WHERE id = ANY($1::int[]);