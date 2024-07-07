BEGIN;

CREATE TABLE containers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE versions (
    id SERIAL PRIMARY KEY,
    container_id INT NOT NULL,
    name VARCHAR(64) NOT NULL,
    is_published BOOLEAN NOT NULL
);

CREATE TABLE blobs (
    id SERIAL PRIMARY KEY,
    checksum CHAR(64) NOT NULL,
    size INT NOT NULL,
    mime_type VARCHAR(64) NOT NULL
);

CREATE TABLE objects (
    id SERIAL PRIMARY KEY,
    version_id INT NOT NULL,
    key VARCHAR(255) NOT NULL,
    blob_id INT NOT NULL
);

CREATE UNIQUE INDEX containers_name_key ON containers (name);

ALTER TABLE versions ADD FOREIGN KEY (container_id) REFERENCES containers (id);
CREATE UNIQUE INDEX versions_container_id_name_key ON versions (container_id, name);
CREATE INDEX versions_is_published_idx ON versions (is_published);

ALTER TABLE objects ADD FOREIGN KEY (version_id) REFERENCES versions (id);
CREATE UNIQUE INDEX objects_version_id_key_key ON objects (version_id, key);
CREATE INDEX objects_key_idx ON objects (key);
ALTER TABLE objects ADD FOREIGN KEY (blob_id) REFERENCES blobs (id);

CREATE UNIQUE INDEX blobs_checksum_key ON blobs (checksum);

COMMIT;
