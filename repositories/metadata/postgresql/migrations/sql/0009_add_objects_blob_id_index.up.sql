BEGIN;

CREATE INDEX objects_blob_id_idx ON objects (blob_id);

COMMIT;
