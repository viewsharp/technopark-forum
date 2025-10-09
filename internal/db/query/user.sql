-- name: CreateUser :exec
INSERT INTO users (nickname, fullname, email, about)
VALUES ($1, $2, $3, $4);

-- name: GetUserByNickname :one
SELECT * FROM users WHERE nickname = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserByNickname :one
UPDATE users
SET fullname = COALESCE(sqlc.narg(fullname), fullname),
    email = COALESCE(sqlc.narg(email), email),
    about = COALESCE(sqlc.narg(about), about)
WHERE nickname = sqlc.arg(nickname)
RETURNING id, nickname, fullname, email, about;