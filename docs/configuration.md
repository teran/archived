# Configuration

All of archived components except archived-cli are configured via environment
variables. archived-cli allows to use environment variables for some parameters
which are expected to be global and mostly uses CLI arguments.

## archived-exporter

| Variable     |     Type     | Required | Default value | Description                                     |
|--------------|:------------:|:--------:|---------------|-------------------------------------------------|
| METRICS_ADDR |    string    |    No    | :8081         | Metrics server address to listen                |
| LOG_LEVEL    | logrus.Level |    No    | info          | Log verbosity level                             |
| METADATA_DSN |    string    |   Yes    |               | Metadata database DSN (PostgreSQL only for now) |

## archived-gc

| Variable     |     Type     | Required | Default value | Description                                     |
|--------------|:------------:|:--------:|---------------|-------------------------------------------------|
| LOG_LEVEL    | logrus.Level |    No    | info          | Log verbosity level                             |
| METADATA_DSN |    string    |   Yes    |               | Metadata database DSN (PostgreSQL only for now) |
| DRY_RUN      |     bool     |    No    | true          | Do not perform any actual changes to data       |

## archived-manager

| Variable                   |     Type      | Required | Default value | Description                                                |
|----------------------------|:-------------:|:--------:|---------------|------------------------------------------------------------|
| ADDR                       |    string     |    No    | :8080         | Manager application server address to listen on            |
| METRICS_ADDR               |    string     |    No    | :8081         | Metrics server address to listen on                        |
| LOG_LEVEL                  | logrus.Level  |    No    | info          | Log verbosity level                                        |
| METADATA_DSN               |    string     |   Yes    |               | Metadata database DSN (PostgreSQL only for now)            |
| BLOB_S3_ENDPOINT           |    string     |   Yes    |               | Blob repository S3 endpoint                                |
| BLOB_S3_BUCKET             |    string     |   Yes    |               | Blob repository S3 bucket                                  |
| BLOB_S3_CREATE_BUCKET      |     bool      |    No    | false         | Whether to create bucket if it doesn't exist yet           |
| BLOB_S3_PRESIGNED_LINK_TTL | time.Duration |    No    | 5m            | Presign url TTL (all blobs are served via presigned links) |
| BLOB_S3_ACCESS_KEY_ID      |    string     |   Yes    |               | S3 Access Key ID                                           |
| BLOB_S3_SECRET_KEY         |    string     |   Yes    |               | S3 Secret key                                              |
| BLOB_S3_REGION             |    string     |    No    | default       | S3 region to use                                           |
| BLOB_S3_DISABLE_SSL        |     bool      |    No    | false         | Whether to disable SSL for S3 connections                  |
| BLOB_S3_FORCE_PATH_STYLE   |     bool      |    No    | true          | Whether to use path-style url format for S3 requests       |

## archived-migrator

| Variable     |     Type     | Required | Default value | Description                                     |
|--------------|:------------:|:--------:|---------------|-------------------------------------------------|
| LOG_LEVEL    | logrus.Level |    No    | info          | Log verbosity level                             |
| METADATA_DSN |    string    |   Yes    |               | Metadata database DSN (PostgreSQL only for now) |

## archived-publisher

| Variable                   |     Type      | Required | Default value | Description                                                                                           |
|----------------------------|:-------------:|:--------:|---------------|-------------------------------------------------------------------------------------------------------|
| ADDR                       |    string     |    No    | :8080         | Publisher application server address to listen on                                                     |
| METRICS_ADDR               |    string     |    No    | :8081         | Metrics server address to listen on                                                                   |
| LOG_LEVEL                  | logrus.Level  |    No    | info          | Log verbosity level                                                                                   |
| METADATA_DSN               |    string     |   Yes    |               | Metadata database DSN (PostgreSQL only for now)                                                       |
| MEMCACHE_SERVERS           |   []string    |    No    | empty list    | Comma-separated list of metadata cache memcache servers. Empty list means metadata cache is disabled. |
| MEMCACHE_TTL               | time.Duration |    No    | 60m           | Metadata cache TTL                                                                                    |
| BLOB_S3_ENDPOINT           |    string     |   Yes    |               | Blob repository S3 endpoint                                                                           |
| BLOB_S3_BUCKET             |    string     |   Yes    |               | Blob repository S3 bucket                                                                             |
| BLOB_S3_PRESIGNED_LINK_TTL | time.Duration |    No    | 5m            | Presign url TTL (all blobs are served via presigned links)                                            |
| BLOB_S3_ACCESS_KEY_ID      |    string     |   Yes    |               | S3 Access Key ID                                                                                      |
| BLOB_S3_SECRET_KEY         |    string     |   Yes    |               | S3 Secret key                                                                                         |
| BLOB_S3_REGION             |    string     |    No    | default       | S3 region to use                                                                                      |
| BLOB_S3_DISABLE_SSL        |     bool      |    No    | false         | Whether to disable SSL for S3 connections                                                             |
| BLOB_S3_FORCE_PATH_STYLE   |     bool      |    No    | true          | Whether to use path-style url format for S3 requests                                                  |

## archived-cli

| Variable                    |  Type  |            Required             | Default value                 | Description                                |
|-----------------------------|:------:|:-------------------------------:|-------------------------------|--------------------------------------------|
| ARCHIVED_CLI_DEBUG          |  bool  |               No                | false                         | Enable debug mode                          |
| ARCHIVED_CLI_TRACE          |  bool  |               No                | false                         | Enable trace mode (debug mode on steroids) |
| ARCHIVED_CLI_ENDPOINT       | string | No (if --endpoint is specified) |                               | Manager API endpoint address               |
| ARCHIVED_CLI_STAT_CACHE_DIR | string |               No                | ~/.cache/archived/cli/objects | Stat-cache directory for objects           |
