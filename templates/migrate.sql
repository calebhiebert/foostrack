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

-- Add Tournament created by id
DO $$ 
    BEGIN
        BEGIN
            ALTER TABLE tournaments ADD COLUMN created_by_id VARCHAR(40) NOT NULL REFERENCES users(id);
        EXCEPTION
            WHEN duplicate_column THEN RAISE NOTICE 'column created_by_id already exists in tournaments.';
        END;
    END;
$$;

-- Add Tournament status by id
DO $$ 
    BEGIN
        BEGIN
            ALTER TABLE tournaments ADD COLUMN status VARCHAR(15) NOT NULL DEFAULT 'signup';
        EXCEPTION
            WHEN duplicate_column THEN RAISE NOTICE 'column status already exists in tournaments.';
        END;
    END;
$$;

-- Add Team color
DO $$ 
    BEGIN
        BEGIN
            ALTER TABLE teams ADD COLUMN color CHAR(7) NOT NULL DEFAULT '#FFFFFF';
        EXCEPTION
            WHEN duplicate_column THEN RAISE NOTICE 'column color already exists in teams.';
        END;
    END;
$$;

-- Add Team tournament id
DO $$ 
    BEGIN
        BEGIN
            ALTER TABLE teams ADD COLUMN tournament_id INTEGER NOT NULL REFERENCES tournaments(id);
        EXCEPTION
            WHEN duplicate_column THEN RAISE NOTICE 'column tournament_id already exists in teams.';
        END;
    END;
$$;

-- Add Bracket Position game id
DO $$ 
    BEGIN
        BEGIN
            ALTER TABLE bracket_positions ADD COLUMN game_id INTEGER NOT NULL REFERENCES games(id);
        EXCEPTION
            WHEN duplicate_column THEN RAISE NOTICE 'column game_id already exists in bracket_positions.';
        END;
    END;
$$;

-- Add Team ID to game events
DO $$ 
    BEGIN
        BEGIN
            ALTER TABLE game_events ADD COLUMN team_id INTEGER REFERENCES teams(id);
        EXCEPTION
            WHEN duplicate_column THEN RAISE NOTICE 'column team_id already exists in game_events.';
        END;
    END;
$$;