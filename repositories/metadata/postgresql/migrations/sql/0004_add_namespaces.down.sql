BEGIN;

ALTER TABLE containers
    DROP COLUMN namespace_id;

DROP TABLE namespaces;

COMMIT;
