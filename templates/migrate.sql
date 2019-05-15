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