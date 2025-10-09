-- name: CreateForum :one
INSERT INTO forums (slug, title, user_nn)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetForumBySlug :one
SELECT *
FROM forums
WHERE slug = $1;

-- name: GetForumBySlugLight :one
SELECT slug, title, user_nn 
FROM forums 
WHERE slug = $1;

-- name: CheckForumExists :one
SELECT slug FROM forums WHERE slug = $1;

-- name: IncreasePostsCount :exec
UPDATE forums
SET posts = posts + sqlc.arg(new_posts_count)::INT
WHERE slug = sqlc.arg(slug);
