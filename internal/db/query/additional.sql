-- Forum queries

-- name: GetForumBySlugLight :one
SELECT slug, title, user_nn 
FROM forums 
WHERE slug = $1;

-- Thread queries

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

-- Post queries

-- name: UpdatePostMessage :exec
UPDATE posts 
SET message = $1, isedited = TRUE
WHERE id = $2;

-- Existence check queries

-- name: CheckThreadExistsById :one
SELECT id FROM threads WHERE id = $1;

-- name: CheckThreadExistsBySlug :one
SELECT slug FROM threads WHERE slug = $1;

-- name: CheckForumExists :one
SELECT slug FROM forums WHERE slug = $1;

-- Vote queries

-- name: UpsertVoteById :exec
INSERT INTO votes (thread_id, user_nn, voice) 
VALUES ($1, $2, $3) 
ON CONFLICT ON CONSTRAINT votes_thread_user_unique 
DO UPDATE SET voice = EXCLUDED.voice;

-- name: UpsertVoteBySlug :exec
INSERT INTO votes (thread_id, user_nn, voice) 
VALUES ((SELECT id FROM threads WHERE slug = $1), $2, $3) 
ON CONFLICT ON CONSTRAINT votes_thread_user_unique 
DO UPDATE SET voice = EXCLUDED.voice;
