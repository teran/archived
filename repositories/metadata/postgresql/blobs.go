package postgresql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

func (r *repository) CreateBLOB(ctx context.Context, checksum string, size uint64, mimeType string) error {
	_, err := psql.
		Insert("blobs").
		Columns(
			"checksum",
			"size",
			"mime_type",
		).
		Values(
			checksum,
			size,
			mimeType,
		).
		RunWith(r.db).
		ExecContext(ctx)

	return errors.Wrap(err, "error executing SQL query")
}

func (r *repository) GetBlobKeyByObject(ctx context.Context, container, version, key string) (string, error) {
	row := psql.
		Select("b.checksum AS checksum").
		From("blobs b").
		Join("objects o ON o.blob_id = b.id").
		Join("versions v ON o.version_id = v.id").
		Join("containers c ON v.container_id = c.id").
		Where(sq.Eq{
			"c.name":         container,
			"v.name":         version,
			"o.key":          key,
			"v.is_published": true,
		}).
		RunWith(r.db).
		QueryRowContext(ctx)

	var checksum string
	if err := row.Scan(&checksum); err != nil {
		return "", errors.Wrap(err, "error looking up BLOB")
	}

	return checksum, nil
}
