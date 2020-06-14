BEGIN;

INSERT INTO todos(title, day, month, year, completed, description) VALUES
    ('Todo 1', DEFAULT, DEFAULT, DEFAULT, false, DEFAULT),
    ('Todo 2', '01', '01', '2018', DEFAULT, DEFAULT),
    ('Todo 3', DEFAULT, DEFAULT, DEFAULT, false, DEFAULT);

COMMIT;
