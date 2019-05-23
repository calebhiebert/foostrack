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

CREATE TABLE IF NOT EXISTS tournaments (
  id            SERIAL        PRIMARY KEY     NOT NULL,
  name          VARCHAR(40)   NOT NULL,
  created_by_id VARCHAR(40)   NOT NULL        REFERENCES users(id),
  status        VARCHAR(15)   NOT NULL        DEFAULT 'signup',
  created_at    TIMESTAMPTZ,
  updated_at    TIMESTAMPTZ,
  deleted_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS teams (
  id            SERIAL        PRIMARY KEY     NOT NULL,
  name          VARCHAR(40)   NOT NULL,
  color         CHAR(7)       NOT NULL        DEFAULT '#FFFFFF',
  tournament_id INTEGER       NOT NULL        REFERENCES tournaments(id),
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
  team_id       INTEGER       REFERENCES teams(id),
  team          VARCHAR(10),
  position      VARCHAR(10),
  created_at    TIMESTAMPTZ,
  updated_at    TIMESTAMPTZ,
  deleted_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS tournament_users (
  tournament_id INTEGER       NOT NULL        REFERENCES tournaments(id),
  user_id       VARCHAR(40)   NOT NULL        REFERENCES users(id),
  team_id       INTEGER                       REFERENCES teams(id),
  created_at    TIMESTAMPTZ,
  updated_at    TIMESTAMPTZ,
  deleted_at    TIMESTAMPTZ,

  PRIMARY KEY (tournament_id, user_id)
);

CREATE TABLE IF NOT EXISTS bracket_positions (
  tournament_id INTEGER       NOT NULL        REFERENCES tournaments(id),
  team_id       INTEGER       NOT NULL        REFERENCES teams(id),
  bracket_level INTEGER       NOT NULL        DEFAULT 0,
  bracket_position INTEGER    NOT NULL        DEFAULT 0,
  game_id       INTEGER       REFERENCES games(id),
  created_at    TIMESTAMPTZ,
  updated_at    TIMESTAMPTZ,
  deleted_at    TIMESTAMPTZ,

  PRIMARY KEY (tournament_id, team_id, bracket_level)
);
