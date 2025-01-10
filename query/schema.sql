CREATE TABLE urls (
  id   text PRIMARY KEY NOT NULL,
  url  text NOT NULL
);

CREATE TABLE vanities (
  id   text PRIMARY KEY NOT NULL,
  url  text NOT NULL
);

CREATE TABLE keys (
  id      text PRIMARY KEY NOT NULL,
  hashed  text NOT NULL,
  admin   BOOLEAN
);