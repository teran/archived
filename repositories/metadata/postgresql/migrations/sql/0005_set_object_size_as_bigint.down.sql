BEGIN;

ALTER TABLE blobs
    ALTER COLUMN size TYPE INT;

COMMIT;