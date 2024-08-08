BEGIN;

CREATE INDEX objects_version_id_idx ON objects (version_id);

COMMIT;
