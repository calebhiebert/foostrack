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

-- Add Win Goals column
DO $$ 
    BEGIN
        BEGIN
            ALTER TABLE games ADD COLUMN win_goals INTEGER NOT NULL DEFAULT 10;
        EXCEPTION
            WHEN duplicate_column THEN RAISE NOTICE 'column win_goals already exists in games.';
        END;
    END;
$$;

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

CREATE OR REPLACE VIEW current_positions AS
  WITH positions AS (
    SELECT g.*, ROW_NUMBER() OVER (PARTITION BY game_id, team, position, event_type ORDER BY id DESC) AS rn 
    FROM game_events AS g
  )
  SELECT * FROM positions 
    WHERE rn = 1
      AND event_type = 'ptp';

CREATE OR REPLACE VIEW user_stats AS
SELECT u.id, u.username,
                          (SELECT COUNT(*)
                            FROM (SELECT COUNT(id)
                            FROM current_positions c
                            WHERE c.user_id = u.id
                            GROUP BY c.game_id) a) AS games_played,

                          (SELECT AVG(count)
                             FROM (SELECT COUNT(id) AS count
                                     FROM game_events g
                                     WHERE g.user_id = u.id
                                     GROUP BY g.game_id) a) AS avg_goals_per_game
  FROM users u;

CREATE OR REPLACE VIEW game_extended AS
  SELECT *, 
      (SELECT COUNT(id) FROM game_events WHERE game_id = g.id AND event_type = 'goal' AND team = 'blue') AS blue_goals,
      (SELECT COUNT(id) FROM game_events WHERE game_id = g.id AND event_type = 'goal' AND team = 'red') AS red_goals,
      (SELECT COUNT(id) FROM game_events WHERE game_id = g.id AND event_type = 'oob') AS oob,
      (SELECT COUNT(id) FROM game_events WHERE game_id = g.id AND event_type = 'dead') AS dead_balls,
      (SELECT created_at FROM game_events ge WHERE ge.game_id = g.id AND ge.event_type = 'start') AS start_time,
      (SELECT created_at FROM game_events ge WHERE ge.game_id = g.id AND ge.event_type = 'end') AS end_time
    FROM games g;