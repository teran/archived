package postgresql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

func (r *repository) CreateBLOB(ctx context.Context, checksum string, size uint64, mimeType string) error {
	_, err := insertQuery(ctx, r.db, psql.
		Insert("blobs").
		Columns(
			"checksum",
			"size",
			"mime_type",
			"created_at",
		).
		Values(
			checksum,
			size,
			mimeType,
			r.tp().UTC(),
		))
	return mapSQLErrors(err)
}

func (r *repository) GetBlobKeyByObject(ctx context.Context, namespace, container, version, key string) (string, error) {
	row, err := selectQueryRow(ctx, r.db, psql.
		Select("b.checksum AS checksum").
		From("blobs b").
		Join("objects o ON o.blob_id = b.id").
		Join("object_keys ok ON ok.id = o.key_id").
		Join("versions v ON o.version_id = v.id").
		Join("containers c ON v.container_id = c.id").
		Join("namespaces ns ON c.namespace_id = ns.id").
		Where(sq.Eq{
			"ns.name":        namespace,
			"c.name":         container,
			"v.name":         version,
			"ok.key":         key,
			"v.is_published": true,
		}))
	if err != nil {
		return "", mapSQLErrors(err)
	}

	var checksum string
	if err := row.Scan(&checksum); err != nil {
		return "", mapSQLErrors(err)
	}

	return checksum, nil
}

func (r *repository) EnsureBlobKey(ctx context.Context, key string, size uint64) error {
	row, err := selectQueryRow(ctx, r.db, psql.
		Select("id").
		From("blobs").
		Where(sq.Eq{
			"checksum": key,
			"size":     size,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var blobID uint
	if err := row.Scan(&blobID); err != nil {
		return mapSQLErrors(err)
	}

	return nil
}
