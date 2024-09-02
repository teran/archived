package postgresql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	log "github.com/sirupsen/logrus"
)

func (r *repository) CreateObject(ctx context.Context, namespace, container, version, key, casKey string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return mapSQLErrors(err)
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("error rolling back")
		}
	}()

	row, err := selectQueryRow(ctx, tx, psql.
		Select("id").
		From("namespaces").
		Where(sq.Eq{"name": namespace}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var namespaceID uint
	if err := row.Scan(&namespaceID); err != nil {
		return mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, tx, psql.
		Select("v.id as id").
		From("containers c").
		Join("versions v ON v.container_id = c.id").
		Where(sq.Eq{
			"c.namespace_id": namespaceID,
			"c.name":         container,
			"v.name":         version,
			"is_published":   false,
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

	row, err = insertQueryRow(ctx, tx, psql.
		Insert("object_keys").
		Columns(
			"key",
			"created_at",
		).
		Values(
			key,
			r.tp().UTC(),
		).
		Suffix("ON CONFLICT (key) DO UPDATE SET key=excluded.key RETURNING id"),
	)
	if err != nil {
		return mapSQLErrors(err)
	}

	var okID uint
	if err := row.Scan(&okID); err != nil {
		return mapSQLErrors(err)
	}

	_, err = insertQuery(ctx, tx, psql.
		Insert("objects").
		Columns(
			"version_id",
			"key_id",
			"blob_id",
			"created_at",
		).
		Values(
			versionID,
			okID,
			blobID,
			r.tp().UTC(),
		))
	if err != nil {
		return mapSQLErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return mapSQLErrors(err)
	}
	return nil
}

func (r *repository) ListObjects(ctx context.Context, namespace, container, version string, offset, limit uint64) (uint64, []string, error) {
	row, err := selectQueryRow(ctx, r.db, psql.
		Select("id").
		From("namespaces").
		Where(sq.Eq{"name": namespace}))
	if err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	var namespaceID uint
	if err := row.Scan(&namespaceID); err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, r.db, psql.
		Select("id").
		From("containers").
		Where(sq.Eq{
			"name":         container,
			"namespace_id": namespaceID,
		}))
	if err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	var containerID uint
	if err := row.Scan(&containerID); err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, r.db, psql.
		Select("id").
		From("versions").
		Where(sq.Eq{
			"container_id": containerID,
			"name":         version,
		}))
	if err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	var versionID uint
	if err := row.Scan(&versionID); err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, r.db, psql.
		Select("COUNT(*)").
		From("objects").
		Where(sq.Eq{"version_id": versionID}))
	if err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	var objectsTotal uint64
	if err := row.Scan(&objectsTotal); err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	rows, err := selectQuery(ctx, r.db, psql.
		Select("ok.key").
		From("object_keys ok").
		Join("objects o ON ok.id = o.key_id").
		Where(sq.Eq{
			"version_id": versionID,
		}).
		OrderBy("ok.key").
		Offset(offset).
		Limit(limit))
	if err != nil {
		return 0, nil, mapSQLErrors(err)
	}
	defer rows.Close()

	result := []string{}
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return 0, nil, mapSQLErrors(err)
		}

		result = append(result, r)
	}

	return objectsTotal, result, mapSQLErrors(rows.Err())
}

func (r *repository) DeleteObject(ctx context.Context, namespace, container, version string, key ...string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return mapSQLErrors(err)
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("error rolling back")
		}
	}()

	row, err := selectQueryRow(ctx, tx, psql.
		Select("id").
		From("namespaces").
		Where(sq.Eq{"name": namespace}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var namespaceID uint
	if err := row.Scan(&namespaceID); err != nil {
		return mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, tx, psql.
		Select("v.id").
		From("versions v").
		Join("containers c ON v.container_id = c.id").
		Where(sq.Eq{
			"c.namespace_id": namespaceID,
			"c.name":         container,
			"v.name":         version,
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
		From("object_keys").
		Where(sq.Eq{
			"key": key,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}
	var okID uint
	if err := row.Scan(&okID); err != nil {
		return mapSQLErrors(err)
	}

	_, err = deleteQuery(ctx, tx, psql.
		Delete("objects").
		Where(sq.Eq{
			"version_id": versionID,
			"key_id":     okID,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return mapSQLErrors(err)
	}
	return nil
}

func (r *repository) RemapObject(ctx context.Context, namespace, container, version, key, newCASKey string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return mapSQLErrors(err)
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("error rolling back")
		}
	}()

	row, err := selectQueryRow(ctx, tx, psql.
		Select("id").
		From("namespaces").
		Where(sq.Eq{"name": namespace}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var namespaceID uint
	if err := row.Scan(&namespaceID); err != nil {
		return mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, tx, psql.
		Select("v.id").
		From("versions v").
		Join("containers c ON v.container_id = c.id").
		Where(sq.Eq{
			"c.namespace_id": namespaceID,
			"c.name":         container,
			"v.name":         version,
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
		Where(sq.Eq{"checksum": newCASKey}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var blobID uint
	if err := row.Scan(&blobID); err != nil {
		return mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, tx, psql.
		Select("id").
		From("object_keys").
		Where(sq.Eq{
			"key": key,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var okID uint
	if err := row.Scan(&okID); err != nil {
		return mapSQLErrors(err)
	}

	_, err = updateQuery(ctx, tx, psql.
		Update("objects").
		Set("blob_id", blobID).
		Where(sq.Eq{
			"version_id": versionID,
			"key_id":     okID,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return mapSQLErrors(err)
	}
	return nil
}
