-- Create Users table
CREATE TABLE IF NOT EXISTS users (
  id            VARCHAR(40)   PRIMARY KEY NOT NULL,
  username      VARCHAR(120)  NOT NULL,
  picture_url   TEXT NOT NULL
);

-- Create games table
CREATE TABLE IF NOT EXISTS games (
  id            SERIAL        PRIMARY KEY,
  win_goals     INTEGER       NOT NULL      DEFAULT 10,
  created_at    TIMESTAMPTZ,
  updated_at    TIMESTAMPTZ,
  deleted_at    TIMESTAMPTZ
);

-- Create game events table
CREATE TABLE IF NOT EXISTS game_events (
  id            SERIAL        PRIMARY KEY     NOT NULL,
  game_id       INTEGER       NOT NULL        REFERENCES games(id),
  user_id       VARCHAR(40)                   REFERENCES users(id),
  event_type    VARCHAR(10)   NOT NULL,
  team          VARCHAR(10),
  position      VARCHAR(10),
  created_at    TIMESTAMPTZ,
  updated_at    TIMESTAMPTZ,
  deleted_at    TIMESTAMPTZ
);
