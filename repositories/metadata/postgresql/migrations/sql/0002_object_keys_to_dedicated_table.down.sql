BEGIN;

ALTER TABLE objects
    ADD COLUMN key VARCHAR(255)
;

WITH sub AS (
    SELECT
        id,
        key
    FROM
        object_keys
    ORDER BY id
)
UPDATE objects SET key=sub.key FROM sub WHERE objects.key_id=sub.id;

ALTER TABLE objects
    ALTER COLUMN key SET NOT NULL,
    DROP COLUMN key_id
;

DROP TABLE object_keys;

COMMIT;
