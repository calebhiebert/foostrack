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