CREATE EXTENSION IF NOT EXISTS CITEXT;

----------- users -----------

CREATE TABLE IF NOT EXISTS users (
  nickname CITEXT PRIMARY KEY,
  fullname TEXT   NOT NULL,
  email    CITEXT NOT NULL UNIQUE,
  about    TEXT
);

----------- forums -----------

CREATE TABLE IF NOT EXISTS forums (
  --   id SERIAL PRIMARY KEY,
  slug    CITEXT PRIMARY KEY,
  title   TEXT                          NOT NULL,
  user_nn CITEXT REFERENCES users (nickname) NOT NULL
);

----------- threads -----------

CREATE TABLE IF NOT EXISTS threads (
  id         SERIAL PRIMARY KEY,
  slug       CITEXT UNIQUE,
  created    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  title      TEXT                            NOT NULL,
  message    TEXT,
  --   votes INTEGER,
  user_nn    CITEXT REFERENCES users (nickname)   NOT NULL,
  forum_slug CITEXT REFERENCES forums (slug) NOT NULL
);

----------- posts -----------

CREATE TABLE IF NOT EXISTS posts (
  id        SERIAL PRIMARY KEY,
  created   TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  isEdited  BOOLEAN                  DEFAULT FALSE,
  message   TEXT                            NOT NULL,
  --   votes INTEGER,
  parent_id INTEGER REFERENCES posts (id), -- Adjacency List
  user_nn   CITEXT REFERENCES users (nickname)   NOT NULL,
  thread_id INTEGER REFERENCES threads (id) NOT NULL
);

----------- vote -----------

CREATE TABLE IF NOT EXISTS votes (
  thread_id INTEGER REFERENCES threads (id) NOT NULL,
  user_nn   citext REFERENCES users (nickname)   NOT NULL,
  voice     INTEGER,
  CONSTRAINT votes_thread_user_unique UNIQUE (thread_id, user_nn)
)
