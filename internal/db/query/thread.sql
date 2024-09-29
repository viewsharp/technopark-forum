-- name: GetThreadBySlug :one
SELECT *
FROM threads
WHERE slug = $1;

-- name: GetThreadByID :one
SELECT *
FROM threads
WHERE id = $1;
