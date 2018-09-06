----------- users -----------

CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  nickname VARCHAR(128) NOT NULL UNIQUE,
  fullname VARCHAR(128) NOT NULL,
  email VARCHAR(128) NOT NULL UNIQUE,
  about TEXT
);


----------- forums -----------

CREATE TABLE IF NOT EXISTS forums (
  --   id SERIAL PRIMARY KEY,
  --   posts INTEGER,
  slug VARCHAR(128) PRIMARY KEY,
  --   threads INTEGER,
  title VARCHAR(128) NOT NULL,
  user_id INTEGER REFERENCES users (id)
);


----------- threads -----------

CREATE TABLE IF NOT EXISTS threads (
  id SERIAL PRIMARY KEY,
  slug VARCHAR(128),
  created DATE DEFAULT CURRENT_DATE,
  title VARCHAR(128) NOT NULL,
  message TEXT,
  --   votes INTEGER,
  user_id INTEGER REFERENCES users (id),
  forum_slug VARCHAR(128) REFERENCES forums (slug)
);


----------- posts -----------

CREATE TABLE IF NOT EXISTS posts (
  id SERIAL PRIMARY KEY,
  created DATE DEFAULT CURRENT_DATE,
  title VARCHAR(128) NOT NULL,
  isEdited BOOLEAN DEFAULT FALSE,
  message TEXT NOT NULL,
  --   votes INTEGER,
  parent_id INTEGER REFERENCES posts (id),-- Adjacency List
  user_id INTEGER REFERENCES users (id),
  thread_id INTEGER REFERENCES threads (id)
);
