BEGIN;

ALTER TABLE containers
    DROP COLUMN namespace_id;

DROP TABLE namespaces;

CREATE UNIQUE INDEX containers_name_key ON containers (name);

COMMIT;
