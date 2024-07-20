package postgresql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

func (r *repository) CreateObject(ctx context.Context, container, version, key, casKey string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "error beginning transaction")
	}
	defer tx.Rollback()

	row, err := selectQueryRow(ctx, tx, psql.
		Select("v.id as id").
		From("containers c").
		Join("versions v ON v.container_id = c.id").
		Where(sq.Eq{
			"c.name":       container,
			"v.name":       version,
			"is_published": false,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var versionID uint
	if err := row.Scan(&versionID); err != nil {
		return mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, tx, psql.
		Select("id").
		From("blobs").
		Where(sq.Eq{"checksum": casKey}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var blobID uint
	if err := row.Scan(&blobID); err != nil {
		return mapSQLErrors(err)
	}

	_, err = psql.
		Insert("objects").
		Columns(
			"version_id",
			"key",
			"blob_id",
		).
		Values(
			versionID,
			key,
			blobID,
		).
		RunWith(tx).
		ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "error executing SQL query")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "error committing transaction")
	}
	return nil
}

func (r *repository) ListObjects(ctx context.Context, container, version string, offset, limit uint64) ([]string, error) {
	row, err := selectQueryRow(ctx, r.db, psql.
		Select("id").
		From("containers").
		Where(sq.Eq{
			"name": container,
		}))
	if err != nil {
		return nil, mapSQLErrors(err)
	}

	var containerID uint
	if err := row.Scan(&containerID); err != nil {
		return nil, mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, r.db, psql.
		Select("id").
		From("versions").
		Where(sq.Eq{
			"container_id": containerID,
			"name":         version,
		}))
	if err != nil {
		return nil, mapSQLErrors(err)
	}

	var versionID uint
	if err := row.Scan(&versionID); err != nil {
		return nil, mapSQLErrors(err)
	}

	rows, err := selectQuery(ctx, r.db, psql.
		Select("key").
		From("objects").
		Where(sq.Eq{
			"version_id": versionID,
		}).
		OrderBy("id").
		Offset(offset).
		Limit(limit))
	if err != nil {
		return nil, mapSQLErrors(err)
	}
	defer rows.Close()

	result := []string{}
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, errors.Wrap(err, "error decoding database result")
		}

		result = append(result, r)
	}

	return result, nil
}

func (r *repository) DeleteObject(ctx context.Context, container, version, key string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "error beginning transaction")
	}
	defer tx.Rollback()

	row, err := selectQueryRow(ctx, tx, psql.
		Select("v.id").
		From("versions v").
		Join("containers c ON v.container_id = c.id").
		Where(sq.Eq{
			"c.name": container,
			"v.name": version,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var versionID uint
	if err := row.Scan(&versionID); err != nil {
		return errors.Wrap(err, "error looking up version")
	}

	_, err = psql.
		Delete("objects").
		Where(sq.Eq{
			"version_id": versionID,
			"key":        key,
		}).
		RunWith(tx).
		ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "error executing SQL query")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "error committing transaction")
	}
	return nil
}

func (r *repository) RemapObject(ctx context.Context, container, version, key, newCASKey string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "error beginning transaction")
	}
	defer tx.Rollback()

	row, err := selectQueryRow(ctx, tx, psql.
		Select("v.id").
		From("versions v").
		Join("containers c ON v.container_id = c.id").
		Where(sq.Eq{
			"c.name": container,
			"v.name": version,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var versionID uint
	if err := row.Scan(&versionID); err != nil {
		return errors.Wrap(err, "error looking up version")
	}

	row, err = selectQueryRow(ctx, tx, psql.
		Select("id").
		From("blobs").
		Where(sq.Eq{"checksum": newCASKey}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var blobID uint
	if err := row.Scan(&blobID); err != nil {
		return errors.Wrap(err, "error looking up blob")
	}

	_, err = psql.
		Update("objects").
		Set("blob_id", blobID).
		Where(sq.Eq{
			"version_id": versionID,
			"key":        key,
		}).
		RunWith(tx).
		ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "error executing SQL query")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "error committing transaction")
	}
	return nil
}
