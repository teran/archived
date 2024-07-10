# archived

Cloud native service to store versioned data in space-efficient manner

archived is applicable if you have amount of low-cardinality data to share
with amount of users/systems. Good example of that task: APT/RPM repository.

## Project status & roadmap

archived is under active development and almost everything is a subject
to change. MVP will be available on first release v0.0.1

The following things are going to be implemented in further releases:

* authentication for manage API
* authentication for access API
* garbage collector
* additional metadata repositories (like MongoDB, CockroachDB, etc.)
* additional version creators for CLI (like RPM repo, APR repo, etc.)

## How it works

archived is inspired by `rsync --link-dest` which allowed to store package
mirrors without duplicating data for decades. And now archived makes this
approach unbound from local file systems by using modern era storage services
under the hood like S3.

To do so archived relies on two storages: metadata and CAS.

Metadata is a some kind of database to store all of the things:

* containers - some kind of directories
* versions - immutable version of the data in container
* objects - named data BLOBs with some additional metadata

Good example of metadata storage is a PostgreSQL database.

CAS storage is a BLOB storage which stores the data behind objects.
CAS is actually an acronym means Content Addressed Storage which describes
how exactly it operates: stores BLOBs under content aware unique key (SHA256
is used by default).

Good example of CAS storage is S3.

This approach allows to reduce raw data usage by linking duplicates instead
if storing copies.

## archived components

archived is built with microservice architecture containing the following
components:

* access - HTTP server to allow data listing and fetching
* manage - gRPC API to manage containers, versions and objects
* CLI - CLI application to interact with manage component

## How build the project manually

archived requires the following dependencies to build:

* Go v1.22+ (prior versions not tested)
* goreleaser v2.0+ (prior versions not tested)
* protoc-gen-go v1.34+ (prior versions not tested)
* protoc-gen-go-grpc v1.4 (prior versions not test)
* docker (to build container images, run some tests)

To build the project just:

```shell
go generate ./...
goreleaser build --snapshot --clean
```

To build container images:

```shell
docker-compose build
```

or build them manually by running:

```shell
docker build -f Dockerfile.component .
```

Where component is one of access, manage, migrate, etc.

## Local development

In some cases it's nice and clean to run the while stack locally.
archived has `docker-compose` way to do that from prebuilt images:

```shell
docker-compose up
```

or by running custom build:

```shell
go generate -v ./... && \
goreleaser build --snapshot --clean && \
docker-compose build && \
docker-compose up || docker-compose down
```

Please note `docker-compose down` at the will automatically remove
containers on stop. Please remove it if you don't need such behavior.
