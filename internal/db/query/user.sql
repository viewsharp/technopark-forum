-- name: GetUserByNickname :one
SELECT * FROM users WHERE nickname = $1;