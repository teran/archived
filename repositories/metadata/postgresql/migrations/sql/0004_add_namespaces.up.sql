BEGIN;

CREATE TABLE namespaces (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

ALTER TABLE containers
    ADD COLUMN namespace_id INT;

WITH c AS (
    INSERT INTO namespaces (name, created_at) VALUES ('default', NOW()) RETURNING id
)
UPDATE containers SET namespace_id=c.id FROM c;

ALTER TABLE containers
    ALTER COLUMN namespace_id SET NOT NULL,
    ADD FOREIGN KEY (namespace_id) REFERENCES namespaces (id);

CREATE INDEX namespaces_name_idx ON namespaces (name);
DROP INDEX containers_name_key;
CREATE UNIQUE INDEX containers_namespace_id_name_key ON containers (namespace_id, name);

COMMIT;
