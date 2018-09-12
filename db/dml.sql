---------- /forum/create ----------
INSERT INTO forums (slug, title, user_id)
VALUES (?, ?, ?);
-- if exists forum with equal slug
SELECT slug, title, user_id
FROM forums
       JOIN users ON forums.user_id = users.id
WHERE slug = ?;

---------- /forum/{slug}/details ----------
SELECT 0 AS "posts", slug, 0 AS "threads", title, user_id
FROM forums
       JOIN users ON forums.user_id = users.id
WHERE slug = ?;

---------- /thread/{slug_or_id}/create ----------
INSERT INTO threads (slug, title, message, user_id, forum_slug)
VALUES (?, ?, ?, ?, ?);

---------- /forum/{slug}/threads ----------
SELECT threads.id,
       slug,
       created,
       title,
       message,
       0 AS "votes",
       users.nickname,
       forum_slug
FROM threads
       JOIN users ON threads.user_id = users.id
WHERE forum_slug = ?;

---------- /user/{nickname}/create ----------
INSERT INTO users (nickname, fullname, email, about)
VALUES (?, ?, ?, ?);
-- if exists user with equal nickname or email
SELECT nickname, fullname, email, about
FROM users
WHERE nickname = ?
UNION
SELECT nickname, fullname, email, about
FROM users
WHERE email = ?;

---------- /user/{nickname}/profile ----------
SELECT nickname, fullname, email, about
FROM users
WHERE nickname = ?;

---------- /user/{nickname}/profile ----------
UPDATE users
SET about    = COALESCE(?, about),
    email    = COALESCE(?, email),
    fullname = COALESCE(?, fullname)
WHERE nickname = ?;

---------- posts ---------
INSERT INTO posts (user_id, message, parent_id, thread_id)
VALUES ((SELECT id FROM users WHERE nickname = ?), ?, ?, ?);

---------- votes ----------
INSERT INTO votes (thread_id, user_id, voice)
VALUES (?, ?, ?)
ON CONFLICT ON CONSTRAINT votes_thread_user_unique DO UPDATE SET thread_id = ?, user_id = ?, voice = ?;

SELECT (SELECT count(*) FROM posts
                               JOIN threads ON posts.thread_id = threads.id WHERE threads.forum_slug = ?) AS "posts",
       slug,
       (SELECT count(*) FROM threads WHERE threads.forum_slug = ?)                                        AS "threads",
       title,
       users.nickname
FROM forums
       JOIN users ON forums.user_id = users.id
WHERE slug = ?;

SELECT count(*)
FROM posts
       JOIN threads ON posts.thread_id = threads.id
WHERE threads.forum_slug = ?;

SELECT *
FROM public.threads;


WITH RECURSIVE recursetree (id, parent_id, name) AS (
    SELECT id, parent_id, name FROM posts
    WHERE parent_id = 0
        UNION ALL
        SELECT t.id, t.parent_id, t.name
        FROM tree t
                    JOIN recursetree rt ON rt.id = t.parent_id
    )
SELECT * FROM recursetree;