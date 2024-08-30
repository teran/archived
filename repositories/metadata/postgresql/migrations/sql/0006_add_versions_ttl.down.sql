BEGIN;

ALTER TABLE containers
    DROP COLUMN version_ttl_seconds;

COMMIT;
