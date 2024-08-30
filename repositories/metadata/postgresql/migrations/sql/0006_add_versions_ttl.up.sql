BEGIN;

ALTER TABLE containers
    ADD COLUMN versions_ttl BIGINT DEFAULT -1;
UPDATE containers SET versions_ttl = -1;
ALTER TABLE containers
    ALTER COLUMN versions_ttl SET NOT NULL;

COMMIT;
