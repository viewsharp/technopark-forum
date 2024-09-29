-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users
(
    id       SERIAL NOT NULL UNIQUE,
    nickname citext COLLATE "ucs_basic" PRIMARY KEY,
    fullname TEXT   NOT NULL,
    email    citext NOT NULL UNIQUE,
    about    TEXT
);

-- CREATE INDEX users_nickname ON users USING HASH ( nickname );

CREATE TABLE IF NOT EXISTS forums
(
    slug    citext PRIMARY KEY,
    title   TEXT                               NOT NULL,
    user_nn citext REFERENCES users (nickname) NOT NULL,
    posts   INTEGER DEFAULT 0, -- Denormalization
    threads INTEGER DEFAULT 0  -- Denormalization
);

-- CREATE INDEX forums_slug ON forums USING HASH ( slug );

CREATE TABLE forum_user
(-- Denormalization
    forum_slug citext REFERENCES forums (slug) NOT NULL,
    user_id    INTEGER REFERENCES users (id)   NOT NULL,
    PRIMARY KEY (user_id, forum_slug)
);

-- CREATE INDEX forum_user_forum_slug ON forum_user USING HASH ( forum_slug );

CREATE TABLE IF NOT EXISTS threads
(
    id         SERIAL PRIMARY KEY,
    slug       citext UNIQUE,
    created    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    title      TEXT                               NOT NULL,
    message    TEXT,
    votes      INTEGER                  DEFAULT 0,
    user_nn    citext REFERENCES users (nickname) NOT NULL,
    forum_slug citext REFERENCES forums (slug)    NOT NULL
);

-- CREATE INDEX threads_slug ON threads USING HASH ( slug );

CREATE INDEX threads__forum_created
    ON threads (forum_slug, created);

CREATE OR REPLACE FUNCTION threadinsert()
    RETURNS TRIGGER AS
$BODY$
BEGIN
    INSERT INTO forum_user (forum_slug, user_id)
    VALUES (new.forum_slug, (SELECT id FROM users WHERE nickname = new.user_nn))
    ON CONFLICT DO NOTHING;
    UPDATE forums SET threads = threads + 1 WHERE slug = new.forum_slug;
    RETURN new;
END;
$BODY$
    LANGUAGE plpgsql;

CREATE TRIGGER threadinsert
    AFTER INSERT
    ON threads
    FOR EACH ROW
EXECUTE PROCEDURE threadinsert();

CREATE TABLE IF NOT EXISTS posts
(
    id        SERIAL PRIMARY KEY,
    created   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    isedited  BOOLEAN                  DEFAULT FALSE,
    message   TEXT                               NOT NULL,
    parent_id INTEGER REFERENCES posts (id),
    user_nn   citext REFERENCES users (nickname) NOT NULL,
    thread_id INTEGER REFERENCES threads (id)    NOT NULL,
    path      INTEGER ARRAY
);

CREATE INDEX posts__thread_id_created
    ON posts (thread_id, id, created);

CREATE TABLE IF NOT EXISTS votes
(
    thread_id INTEGER REFERENCES threads (id)    NOT NULL,
    user_nn   citext REFERENCES users (nickname) NOT NULL,
    voice     INTEGER,
    CONSTRAINT votes_thread_user_unique UNIQUE (thread_id, user_nn)
);

CREATE OR REPLACE FUNCTION voteupdate()
    RETURNS TRIGGER AS
$BODY$
BEGIN
    IF old.voice = -1 AND new.voice = 1
    THEN
        UPDATE threads SET votes = votes + 2 WHERE id = new.thread_id;
    END IF;
    IF old.voice = 1 AND new.voice = -1
    THEN
        UPDATE threads SET votes = votes - 2 WHERE id = new.thread_id;
    END IF;
    RETURN new;
END;
$BODY$
    LANGUAGE plpgsql;

CREATE TRIGGER voteupdate
    AFTER UPDATE
    ON votes
    FOR EACH ROW
EXECUTE PROCEDURE voteupdate();

CREATE OR REPLACE FUNCTION voteinsert()
    RETURNS TRIGGER AS
$BODY$
BEGIN
    UPDATE threads SET votes = votes + new.voice WHERE id = new.thread_id;
    RETURN new;
END;
$BODY$
    LANGUAGE plpgsql;

CREATE TRIGGER voteinsert
    AFTER INSERT
    ON votes
    FOR EACH ROW
EXECUTE PROCEDURE voteinsert();


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER voteinsert ON votes;
DROP FUNCTION voteinsert;
DROP TRIGGER voteupdate ON votes;
DROP FUNCTION voteupdate;
DROP TABLE votes;
DROP INDEX posts__thread_id_created;
DROP TABLE posts;
DROP TRIGGER threadinsert ON threads;
DROP FUNCTION threadinsert;
DROP TABLE forum_user;
DROP TABLE forums;
DROP TABLE users;

-- +goose StatementEnd
