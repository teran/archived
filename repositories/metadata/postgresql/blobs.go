package postgresql

import (
	"context"

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
