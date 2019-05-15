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

-- Create teams table
CREATE TABLE IF NOT EXISTS teams (
  id            SERIAL        PRIMARY KEY,
  name          VARCHAR(40)   NOT NULL,
  created_at    TIMESTAMPTZ,
  updated_at    TIMESTAMPTZ,
  deleted_at    TIMESTAMPTZ
);

-- Create team_users table
CREATE TABLE IF NOT EXISTS team_users (
  team_id       INTEGER       NOT NULL,
  user_id       VARCHAR(40)   NOT NULL,
  created_at    TIMESTAMPTZ,
  updated_at    TIMESTAMPTZ,
  deleted_at    TIMESTAMPTZ,

  PRIMARY KEY (team_id, user_id)
);
