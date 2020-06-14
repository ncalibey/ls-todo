BEGIN;

-- This removes all data from the table and resets the `id` sequence
-- (i.e., the first data inserted will have an id of `1`).
TRUNCATE TABLE todos RESTART IDENTITY;

COMMIT;
