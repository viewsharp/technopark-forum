-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS CITEXT;

CREATE TABLE IF NOT EXISTS users
(
    id       SERIAL NOT NULL UNIQUE,
    nickname CITEXT COLLATE "ucs_basic" PRIMARY KEY,
    fullname TEXT   NOT NULL,
    email    CITEXT NOT NULL UNIQUE,
    about    TEXT
);

CREATE INDEX users_nickname ON users USING HASH ( nickname );

CREATE TABLE IF NOT EXISTS forums
(
    slug    CITEXT PRIMARY KEY,
    title   TEXT                               NOT NULL,
    user_nn CITEXT REFERENCES users (nickname) NOT NULL,
    posts   INTEGER DEFAULT 0, -- Denormalization
    threads INTEGER DEFAULT 0  -- Denormalization
);

CREATE INDEX forums_slug ON forums USING HASH ( slug );

CREATE TABLE forum_user
(-- Denormalization
    forum_slug CITEXT,
    user_id    INTEGER,
    PRIMARY KEY (user_id, forum_slug)
);

CREATE INDEX forum_user_forum_slug ON forum_user USING HASH ( forum_slug );

CREATE TABLE IF NOT EXISTS threads
(
    id         SERIAL PRIMARY KEY,
    slug       CITEXT UNIQUE,
    created    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    title      TEXT                               NOT NULL,
    message    TEXT,
    votes      INTEGER                  DEFAULT 0,
    user_nn    CITEXT REFERENCES users (nickname) NOT NULL,
    forum_slug CITEXT REFERENCES forums (slug)    NOT NULL
);

CREATE INDEX threads_slug ON threads USING HASH ( slug );

CREATE INDEX threads__forum_created
    ON threads (forum_slug, created);

CREATE OR REPLACE FUNCTION threadInsert()
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

CREATE TRIGGER threadInsert
    AFTER INSERT
    ON threads
    FOR EACH ROW
EXECUTE PROCEDURE threadInsert();

CREATE TABLE IF NOT EXISTS posts
(
    id        SERIAL PRIMARY KEY,
    created   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    isEdited  BOOLEAN                  DEFAULT FALSE,
    message   TEXT                               NOT NULL,
    parent_id INTEGER REFERENCES posts (id),
    user_nn   CITEXT REFERENCES users (nickname) NOT NULL,
    thread_id INTEGER REFERENCES threads (id)    NOT NULL,
    path      INTEGER ARRAY
);

CREATE INDEX posts__thread_id_created
    ON posts (thread_id, id, created);

CREATE TABLE IF NOT EXISTS votes
(
    thread_id INTEGER REFERENCES threads (id)    NOT NULL,
    user_nn   CITEXT REFERENCES users (nickname) NOT NULL,
    voice     INTEGER,
    CONSTRAINT votes_thread_user_unique UNIQUE (thread_id, user_nn)
);

CREATE OR REPLACE FUNCTION voteUpdate()
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

CREATE TRIGGER voteUpdate
    AFTER UPDATE
    ON votes
    FOR EACH ROW
EXECUTE PROCEDURE voteUpdate();

CREATE OR REPLACE FUNCTION voteInsert()
    RETURNS TRIGGER AS
$BODY$
BEGIN
    UPDATE threads SET votes = votes + new.voice WHERE id = new.thread_id;
    RETURN new;
END;
$BODY$
    LANGUAGE plpgsql;

CREATE TRIGGER voteInsert
    AFTER INSERT
    ON votes
    FOR EACH ROW
EXECUTE PROCEDURE voteInsert();


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER voteInsert ON votes;
DROP FUNCTION voteInsert;
DROP TRIGGER voteUpdate ON votes;
DROP FUNCTION voteUpdate;
DROP TABLE votes;
DROP INDEX posts__thread_id_created;
DROP TABLE posts;
DROP TRIGGER threadInsert on threads;
DROP FUNCTION threadInsert;
DROP TABLE forum_user;
DROP TABLE forums;
DROP TABLE users;

-- +goose StatementEnd
