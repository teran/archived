BEGIN;

ALTER TABLE blobs
    ADD CONSTRAINT size_range CHECK (size >= 0)
;

ALTER TABLE containers
    ADD CONSTRAINT version_ttl_seconds_range CHECK (version_ttl_seconds >= -1)
;

COMMIT;
