BEGIN;

ALTER TABLE blobs
    DROP CONSTRAINT size_range
;

ALTER TABLE containers
    DROP CONSTRAINT version_ttl_seconds_range
;

COMMIT;
