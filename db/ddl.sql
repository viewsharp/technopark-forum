CREATE EXTENSION IF NOT EXISTS CITEXT;

--------------------------------------- users ---------------------------------------

CREATE TABLE IF NOT EXISTS users (
  id       SERIAL,
  nickname CITEXT COLLATE "ucs_basic" PRIMARY KEY,
  fullname TEXT   NOT NULL,
  email    CITEXT NOT NULL UNIQUE,
  about    TEXT
);

CREATE INDEX users__id ON users(id);

--------------------------------------- forums ---------------------------------------

CREATE TABLE IF NOT EXISTS forums (
  slug    CITEXT PRIMARY KEY,
  title   TEXT                               NOT NULL,
  user_nn CITEXT REFERENCES users (nickname) NOT NULL,
  posts   INTEGER DEFAULT 0, -- Denormalization
  threads INTEGER DEFAULT 0 -- Denormalization
);

--------------------------------------- forum_user ---------------------------------------

CREATE TABLE forum_user (-- Denormalization
  forum_slug CITEXT,
  user_id    INTEGER,
  PRIMARY KEY (user_id, forum_slug)
);

--------------------------------------- threads ---------------------------------------

CREATE TABLE IF NOT EXISTS threads (
  id         SERIAL PRIMARY KEY,
  slug       CITEXT UNIQUE,
  created    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  title      TEXT                                 NOT NULL,
  message    TEXT,
  votes      INTEGER                  DEFAULT 0,
  user_nn    CITEXT REFERENCES users (nickname)   NOT NULL,
  forum_slug CITEXT REFERENCES forums (slug)      NOT NULL
);

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
  FOR EACH ROW EXECUTE PROCEDURE threadInsert();

--------------------------------------- posts ---------------------------------------

CREATE TABLE IF NOT EXISTS posts (
  id        SERIAL PRIMARY KEY,
  created   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  isEdited  BOOLEAN                  DEFAULT FALSE,
  message   TEXT                                 NOT NULL,
  parent_id INTEGER REFERENCES posts (id),
  user_nn   CITEXT REFERENCES users (nickname)   NOT NULL,
  thread_id INTEGER REFERENCES threads (id)      NOT NULL,
  path      INTEGER ARRAY
);

CREATE INDEX posts__thread_id_created
  ON posts (thread_id, id, created);

--------------------------------------- vote ---------------------------------------

CREATE TABLE IF NOT EXISTS votes (
  thread_id INTEGER REFERENCES threads (id)      NOT NULL,
  user_nn   citext REFERENCES users (nickname)   NOT NULL,
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
  FOR EACH ROW EXECUTE PROCEDURE voteUpdate();

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
  FOR EACH ROW EXECUTE PROCEDURE voteInsert();

