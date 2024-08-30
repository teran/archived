BEGIN;

ALTER TABLE containers
    DROP COLUMN versions_ttl;

COMMIT;
