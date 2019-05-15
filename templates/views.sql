CREATE OR REPLACE VIEW current_positions AS
  WITH positions AS (
    SELECT g.*, ROW_NUMBER() OVER (PARTITION BY game_id, team, position, event_type ORDER BY id DESC) AS rn 
    FROM game_events AS g
    WHERE g.deleted_at IS NULL
  )
  SELECT * FROM positions 
    WHERE rn = 1
      AND event_type = 'ptp';

CREATE OR REPLACE VIEW game_extended AS
  SELECT *, 
      (SELECT COUNT(id) FROM game_events WHERE game_id = g.id 
        AND (
          (event_type = 'goal' AND team = 'blue' AND deleted_at IS NULL) OR (event_type = 'antigoal' AND team = 'red' AND deleted_at IS NULL)
          )) AS blue_goals,
      (SELECT COUNT(id) FROM game_events WHERE game_id = g.id 
        AND (
          (event_type = 'goal' AND team = 'red' AND deleted_at IS NULL) OR (event_type = 'antigoal' AND team = 'blue' AND deleted_at IS NULL)
        )) AS red_goals,
      (SELECT COUNT(id) FROM game_events WHERE game_id = g.id AND event_type = 'oob' AND deleted_at IS NULL) AS oob,
      (SELECT COUNT(id) FROM game_events WHERE game_id = g.id AND event_type = 'dead' AND deleted_at IS NULL) AS dead_balls,
      (SELECT created_at FROM game_events ge WHERE ge.game_id = g.id AND ge.event_type = 'start') AS start_time,
      (SELECT created_at FROM game_events ge WHERE ge.game_id = g.id AND ge.event_type = 'end') AS end_time,
      (SELECT ARRAY_AGG(cp.user_id) AS blue_members FROM current_positions cp  WHERE cp.game_id = g.id AND cp.team = 'blue' GROUP BY cp.game_id),
      (SELECT ARRAY_AGG(cp.user_id) AS red_members FROM current_positions cp  WHERE cp.game_id = g.id AND cp.team = 'red' GROUP BY cp.game_id)
    FROM games g;

DROP VIEW IF EXISTS user_stats;
DROP VIEW IF EXISTS goals;

CREATE OR REPLACE VIEW goals AS
SELECT g.*, (SELECT user_id
             FROM current_positions c
             WHERE c.game_id = g.game_id
               AND c.position = 'goalie'
               AND c.created_at < g.created_at
               AND c.team <> g.team
             ORDER BY c.id DESC
             LIMIT 1) AS goalie_id
FROM game_events g
WHERE g.event_type = 'goal' AND g.deleted_at IS NULL;

CREATE OR REPLACE VIEW user_stats AS
SELECT u.*,
                          (SELECT COUNT(*)
                            FROM (SELECT COUNT(id)
                            FROM current_positions c
                            WHERE c.user_id = u.id
                            GROUP BY c.game_id) a) AS games_played,

                          (SELECT COUNT(*)
                            FROM (SELECT COUNT(id)
                            FROM current_positions c
                            WHERE c.user_id = u.id AND c.team = 'red'
                            GROUP BY c.game_id) a) AS games_played_red,

                          (SELECT COUNT(*)
                            FROM (SELECT COUNT(id)
                            FROM current_positions c
                            WHERE c.user_id = u.id AND c.team = 'blue'
                            GROUP BY c.game_id) a) AS games_played_blue,

                          (SELECT AVG(count)
                             FROM (SELECT COUNT(id) AS count
                                     FROM game_events g
                                     WHERE g.user_id = u.id AND g.deleted_at IS NULL
                                     GROUP BY g.game_id) a) AS avg_goals_per_game,

                          (SELECT COUNT(id)
                                     FROM game_events g
                                     WHERE g.user_id = u.id AND g.deleted_at IS NULL
                                       AND g.event_type = 'antigoal') AS antigoals,

                          (SELECT COUNT(g.id)
                             FROM game_extended g
                               JOIN current_positions cp ON g.id = cp.game_id
                             WHERE cp.user_id = u.id
                               AND ((cp.team = 'blue' AND g.win_goals = g.blue_goals)
                               OR (cp.team = 'red' AND g.win_goals = g.red_goals))) AS games_won,

                          (SELECT COUNT(id)
                             FROM game_events g
                             WHERE g.user_id = u.id AND g.deleted_at IS NULL
                                AND g.event_type = 'goal') AS goals,

                          (SELECT COUNT(id)
                            FROM goals g
                            WHERE g.goalie_id = u.id) AS non_saves,
                                                        
                          (SELECT COUNT(*)
                            FROM (SELECT COUNT(id)
                            FROM current_positions c
                            WHERE c.user_id = u.id AND c.position = 'goalie'
                            GROUP BY c.game_id) a) AS games_played_goalie,

                          (SELECT COUNT(*)
                            FROM (SELECT COUNT(id)
                            FROM current_positions c
                            WHERE c.user_id = u.id AND c.position = 'forward'
                            GROUP BY c.game_id) a) AS games_played_forward
  FROM users u;