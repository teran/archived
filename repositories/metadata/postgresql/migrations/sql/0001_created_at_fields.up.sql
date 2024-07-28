BEGIN;

ALTER TABLE containers
    ADD COLUMN created_at TIMESTAMP;
UPDATE containers SET created_at = NOW();
ALTER TABLE containers
    ALTER COLUMN created_at SET NOT NULL;

ALTER TABLE versions
    ADD COLUMN created_at TIMESTAMP;
UPDATE versions SET created_at = NOW();
ALTER TABLE versions
    ALTER COLUMN created_at SET NOT NULL;

ALTER TABLE objects
    ADD COLUMN created_at TIMESTAMP;
UPDATE objects SET created_at = NOW();
ALTER TABLE objects
    ALTER COLUMN created_at SET NOT NULL;

ALTER TABLE blobs
    ADD COLUMN created_at TIMESTAMP;
UPDATE blobs SET created_at = NOW();
ALTER TABLE blobs
    ALTER COLUMN created_at SET NOT NULL;

COMMIT;
