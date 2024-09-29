-- name: CreateForumUser :batchexec
INSERT INTO forum_user (forum_slug, user_id)
VALUES ($1, (SELECT id FROM public.users WHERE users.nickname = $2))
ON CONFLICT DO NOTHING;