BEGIN;

ALTER TABLE containers
    DROP COLUMN created_at
;

ALTER TABLE versions
    DROP COLUMN created_at
;

ALTER TABLE objects
    DROP COLUMN created_at
;

ALTER TABLE blobs
    DROP COLUMN created_at
;

COMMIT;
