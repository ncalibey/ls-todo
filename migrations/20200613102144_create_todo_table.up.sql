BEGIN;

CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    title TEXT DEFAULT '' NOT NULL,
    description TEXT DEFAULT '' NOT NULL,
    -- Note that normally we wouldn't store dates as text; this is just for simplicities
    -- sake.
    day TEXT DEFAULT '' NOT NULL,
    month TEXT DEFAULT '' NOT NULL,
    year TEXT DEFAULT '' NOT NULL,
    completed BOOL DEFAULT 'f' NOT NULL
);

COMMIT;
