BEGIN;

CREATE TABLE object_keys (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

ALTER TABLE objects
    ADD COLUMN key_id INT
;

WITH sub AS (
    INSERT INTO object_keys (key, created_at)
        SELECT
            DISTINCT(key) AS key,
            MAX(created_at) AS created_at
        FROM
            objects
        GROUP BY key
    RETURNING id, key
)
UPDATE objects SET key_id=sub.id FROM sub WHERE objects.key=sub.key;

ALTER TABLE objects
    ALTER COLUMN key_id SET NOT NULL,
    DROP COLUMN key
;

ALTER TABLE objects ADD FOREIGN KEY (key_id) REFERENCES object_keys (id);
CREATE UNIQUE INDEX object_keys_key_key ON object_keys (key);
CREATE UNIQUE INDEX objects_version_id_key_id_key ON objects (version_id, key_id);

COMMIT;
