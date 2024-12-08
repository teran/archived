package postgresql

import (
	"context"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	log "github.com/sirupsen/logrus"
	ptr "github.com/teran/go-ptr"

	"github.com/teran/archived/models"
	"github.com/teran/archived/repositories/metadata"
)

const (
	defaultLimit             uint64 = 1000
	expiredVersionsBatchSize int    = 1000
)

func (r *repository) CreateVersion(ctx context.Context, namespace, container string) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", mapSQLErrors(err)
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
		return "", mapSQLErrors(err)
	}

	var namespaceID uint
	if err := row.Scan(&namespaceID); err != nil {
		return "", mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, tx, psql.
		Select("id").
		From("containers").
		Where(sq.Eq{
			"namespace_id": namespaceID,
			"name":         container,
		}))
	if err != nil {
		return "", mapSQLErrors(err)
	}

	var containerID uint
	if err := row.Scan(&containerID); err != nil {
		return "", metadata.ErrNotFound
	}

	versionTimestamp := r.tp().UTC()
	versionID := versionTimestamp.Format("20060102150405")

	_, err = insertQuery(ctx, tx, psql.
		Insert("versions").
		Columns(
			"container_id",
			"name",
			"is_published",
			"created_at",
		).
		Values(
			containerID,
			versionID,
			false,
			versionTimestamp,
		))
	if err != nil {
		return "", mapSQLErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return "", mapSQLErrors(err)
	}

	return versionID, nil
}

func (r *repository) GetLatestPublishedVersionByContainer(ctx context.Context, namespace, container string) (string, error) {
	row, err := selectQueryRow(ctx, r.db, psql.
		Select("id").
		From("namespaces").
		Where(sq.Eq{"name": namespace}))
	if err != nil {
		return "", mapSQLErrors(err)
	}

	var namespaceID uint
	if err := row.Scan(&namespaceID); err != nil {
		return "", mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, r.db, psql.
		Select("v.name").
		From("versions v").
		Join("containers c ON v.container_id = c.id").
		Where(sq.Eq{
			"c.namespace_id": namespaceID,
			"c.name":         container,
			"v.is_published": true,
		}).
		OrderBy("v.created_at DESC").
		Limit(1),
	)
	if err != nil {
		return "", mapSQLErrors(err)
	}

	var versionName string
	if err := row.Scan(&versionName); err != nil {
		return "", mapSQLErrors(err)
	}

	return versionName, nil
}

func (r *repository) ListPublishedVersionsByContainer(ctx context.Context, namespace, container string) ([]models.Version, error) {
	_, versions, err := r.listVersionsByContainer(ctx, namespace, container, ptr.Bool(true), 0, 0)
	return versions, err
}

func (r *repository) ListAllVersionsByContainer(ctx context.Context, namespace, container string) ([]models.Version, error) {
	_, versions, err := r.listVersionsByContainer(ctx, namespace, container, nil, 0, 0)
	return versions, err
}

func (r *repository) ListUnpublishedVersionsByContainer(ctx context.Context, namespace, container string) ([]models.Version, error) {
	_, versions, err := r.listVersionsByContainer(ctx, namespace, container, ptr.Bool(false), 0, 0)
	return versions, err
}

func (r *repository) ListPublishedVersionsByContainerAndPage(ctx context.Context, namespace, container string, offset, limit uint64) (uint64, []models.Version, error) {
	return r.listVersionsByContainer(ctx, namespace, container, ptr.Bool(true), offset, limit)
}

func (r *repository) listVersionsByContainer(ctx context.Context, namespace, container string, isPublished *bool, offset, limit uint64) (uint64, []models.Version, error) {
	if limit == 0 {
		limit = defaultLimit
	}

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
			"namespace_id": namespaceID,
			"name":         container,
		}))
	if err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	var containerID uint64
	if err := row.Scan(&containerID); err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	condition := sq.Eq{
		"container_id": containerID,
	}

	if isPublished != nil {
		condition["is_published"] = *isPublished
	}

	row, err = selectQueryRow(ctx, r.db, psql.
		Select("COUNT(*)").
		From("versions").
		Where(condition))
	if err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	var versionsTotal uint64
	if err := row.Scan(&versionsTotal); err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	rows, err := selectQuery(ctx, r.db, psql.
		Select("name", "is_published", "created_at").
		From("versions").
		Where(condition).
		OrderBy("created_at DESC").
		Offset(offset).
		Limit(limit))
	if err != nil {
		return 0, nil, mapSQLErrors(err)
	}
	defer rows.Close()

	result := []models.Version{}
	for rows.Next() {
		var (
			r         models.Version
			createdAt time.Time
		)

		if err := rows.Scan(&r.Name, &r.IsPublished, &createdAt); err != nil {
			return 0, nil, mapSQLErrors(err)
		}
		r.CreatedAt = time.Date(
			createdAt.Year(), createdAt.Month(), createdAt.Day(),
			createdAt.Hour(), createdAt.Minute(), createdAt.Second(), createdAt.Nanosecond(),
			time.UTC,
		)

		result = append(result, r)
	}

	return versionsTotal, result, mapSQLErrors(rows.Err())
}

func (r *repository) MarkVersionPublished(ctx context.Context, namespace, container, version string) error {
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
		Select("id").
		From("containers").
		Where(sq.Eq{
			"namespace_id": namespaceID,
			"name":         container,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var containerID uint
	if err := row.Scan(&containerID); err != nil {
		return metadata.ErrNotFound
	}

	_, err = updateQuery(ctx, tx, psql.
		Update("versions").
		Set("is_published", true).
		Where(sq.Eq{
			"container_id": containerID,
			"name":         version,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return mapSQLErrors(err)
	}

	return nil
}

func (r *repository) DeleteVersion(ctx context.Context, namespace, container, version string) error {
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

	var versionID uint64
	if err := row.Scan(&versionID); err != nil {
		return mapSQLErrors(err)
	}

	_, err = deleteQuery(ctx, tx, psql.
		Delete("objects").
		Where(sq.Eq{
			"version_id": versionID,
		}),
	)
	if err != nil {
		return mapSQLErrors(err)
	}

	_, err = deleteQuery(ctx, tx, psql.
		Delete("versions").
		Where(sq.Eq{
			"id": versionID,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return mapSQLErrors(err)
	}
	return nil
}

func (r *repository) DeleteExpiredVersionsWithObjects(ctx context.Context, unpublishedVersionsMaxAge time.Duration) error {
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

	now := r.tp().UTC()

	q := psql.
		Select(
			"v.id AS version_id",
		).
		From("containers c").
		Join("versions v ON v.container_id = c.id").
		Join("namespaces n ON n.id = c.namespace_id").
		Where(
			sq.Or{
				sq.And{
					sq.Gt{
						"c.version_ttl_seconds": 0,
					},
					sq.Expr("v.created_at <= (?::timestamp - c.version_ttl_seconds * interval '1 second')", now.Format(time.RFC3339)),
				},
				sq.And{
					sq.Eq{
						"v.is_published": false,
					},
					sq.Expr("v.created_at <= ?::timestamp", now.Add(-1*unpublishedVersionsMaxAge).Format(time.RFC3339)),
				},
			},
		)

	rows, err := selectQuery(ctx, tx, q)
	if err != nil {
		return mapSQLErrors(err)
	}
	defer rows.Close()

	deleteCandidates := []uint64{}
	for rows.Next() {
		var versionID uint64
		if err := rows.Scan(&versionID); err != nil {
			return mapSQLErrors(err)
		}

		deleteCandidates = append(deleteCandidates, versionID)
	}

	if err := rows.Err(); err != nil {
		return mapSQLErrors(err)
	}

	//
	// lib/pq (and probably PostgreSQL itself) has a limit of 65k arguments so let's batch 'em
	//
	if err := indexChunks(len(deleteCandidates), expiredVersionsBatchSize, func(start, end int) error {
		if _, err := deleteQuery(ctx, tx, psql.
			Delete("objects").
			Where(sq.Eq{
				"version_id": deleteCandidates[start:end],
			}),
		); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return mapSQLErrors(err)
	}

	if err := indexChunks(len(deleteCandidates), expiredVersionsBatchSize, func(start, end int) error {
		if _, err := deleteQuery(ctx, tx, psql.
			Delete("versions").
			Where(sq.Eq{
				"id": deleteCandidates[start:end],
			}),
		); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return mapSQLErrors(err)
	}

	orphanedObjectKeyIDs := []uint64{}
	rows, err = selectQuery(ctx, tx, psql.
		Select(
			"ok.id AS id",
		).
		From("object_keys ok").
		LeftJoin("objects o ON o.key_id = ok.id").
		Where(sq.Eq{
			"o.key_id": nil,
		}),
	)
	if err != nil {
		return mapSQLErrors(err)
	}
	defer rows.Close()

	for rows.Next() {
		var keyID uint64
		if err := rows.Scan(&keyID); err != nil {
			return mapSQLErrors(err)
		}

		orphanedObjectKeyIDs = append(orphanedObjectKeyIDs, keyID)
	}

	if err := indexChunks(len(orphanedObjectKeyIDs), expiredVersionsBatchSize, func(start, end int) error {
		if _, err := deleteQuery(ctx, tx, psql.
			Delete("object_keys").
			Where(sq.Eq{
				"id": orphanedObjectKeyIDs[start:end],
			}),
		); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return mapSQLErrors(err)
	}

	if err := rows.Err(); err != nil {
		return mapSQLErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return mapSQLErrors(err)
	}
	return nil
}

func indexChunks(length, chuckLen int, fn func(start, end int) error) error {
	if chuckLen <= 0 {
		return errors.ErrUnsupported
	}

	for i := 0; i < length; i += chuckLen {
		l := minInt(i+chuckLen, length)
		err := fn(i, l)
		if err != nil {
			return err
		}
	}
	return nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
