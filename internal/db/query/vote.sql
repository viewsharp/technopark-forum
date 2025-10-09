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

