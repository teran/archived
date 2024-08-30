BEGIN;

ALTER TABLE containers
    ADD COLUMN version_ttl_seconds BIGINT DEFAULT -1;
UPDATE containers SET version_ttl_seconds = -1;
ALTER TABLE containers
    ALTER COLUMN version_ttl_seconds SET NOT NULL;

COMMIT;
