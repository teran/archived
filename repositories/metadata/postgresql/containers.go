package postgresql

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/models"
)

func (r *repository) CreateContainer(ctx context.Context, namespace, name string) error {
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

	_, err = insertQuery(ctx, tx, psql.
		Insert("containers").
		Columns(
			"name",
			"namespace_id",
			"created_at",
		).
		Values(
			name,
			namespaceID,
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

func (r *repository) RenameContainer(ctx context.Context, namespace, oldName, newNamespace, newName string) error {
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
		From("namespaces").
		Where(sq.Eq{"name": newNamespace}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var newNamespaceID uint
	if err := row.Scan(&newNamespaceID); err != nil {
		return mapSQLErrors(err)
	}

	row, err = selectQueryRow(ctx, tx, psql.
		Select("id").
		From("containers").
		Where(sq.Eq{
			"name":         oldName,
			"namespace_id": namespaceID,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var containerID uint
	if err := row.Scan(&containerID); err != nil {
		return mapSQLErrors(err)
	}

	_, err = updateQuery(ctx, tx, psql.
		Update("containers").
		Set("name", newName).
		Set("namespace_id", newNamespaceID).
		Where(sq.Eq{
			"id": containerID,
		}),
	)
	if err != nil {
		return mapSQLErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return mapSQLErrors(err)
	}
	return nil
}

func (r *repository) SetContainerVersionsTTL(ctx context.Context, namespace, name string, ttl time.Duration) error {
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
			"name":         name,
			"namespace_id": namespaceID,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var containerID uint
	if err := row.Scan(&containerID); err != nil {
		return mapSQLErrors(err)
	}

	_, err = updateQuery(ctx, tx, psql.
		Update("containers").
		Set("version_ttl_seconds", ttl.Seconds()).
		Where(sq.Eq{
			"name":         name,
			"namespace_id": namespaceID,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return mapSQLErrors(err)
	}
	return nil
}

func (r *repository) ListContainers(ctx context.Context, namespace string) ([]models.Container, error) {
	row, err := selectQueryRow(ctx, r.db, psql.
		Select("id").
		From("namespaces").
		Where(sq.Eq{"name": namespace}))
	if err != nil {
		return nil, mapSQLErrors(err)
	}

	var namespaceID uint
	if err := row.Scan(&namespaceID); err != nil {
		return nil, mapSQLErrors(err)
	}

	rows, err := selectQuery(ctx, r.db, psql.
		Select("name", "created_at", "version_ttl_seconds").
		From("containers").
		Where(sq.Eq{
			"namespace_id": namespaceID,
		}).
		OrderBy("name"))
	if err != nil {
		return nil, mapSQLErrors(err)
	}
	defer rows.Close()

	result := []models.Container{}
	for rows.Next() {
		var (
			r         models.Container
			createdAt time.Time
		)
		if err := rows.Scan(&r.Name, &createdAt, &r.VersionsTTL); err != nil {
			return nil, mapSQLErrors(err)
		}
		r.CreatedAt = time.Date(
			createdAt.Year(), createdAt.Month(), createdAt.Day(),
			createdAt.Hour(), createdAt.Minute(), createdAt.Second(), createdAt.Nanosecond(),
			time.UTC,
		)

		result = append(result, r)
	}

	return result, mapSQLErrors(rows.Err())
}

func (r *repository) ListContainersByPage(ctx context.Context, namespace string, offset, limit uint64) (uint64, []models.Container, error) {
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
		Select("COUNT(*)").
		From("containers").
		Where(sq.Eq{
			"namespace_id": namespaceID,
		}))
	if err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	var containersTotal uint64
	if err := row.Scan(&containersTotal); err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	rows, err := selectQuery(ctx, r.db, psql.
		Select("name", "created_at", "version_ttl_seconds").
		From("containers").
		Where(sq.Eq{
			"namespace_id": namespaceID,
		}).
		OrderBy("name").
		Limit(limit).
		Offset(offset))
	if err != nil {
		return 0, nil, mapSQLErrors(err)
	}
	defer rows.Close()

	result := []models.Container{}
	for rows.Next() {
		var (
			r         models.Container
			createdAt time.Time
		)
		if err := rows.Scan(&r.Name, &createdAt, &r.VersionsTTL); err != nil {
			return 0, nil, mapSQLErrors(err)
		}
		r.CreatedAt = time.Date(
			createdAt.Year(), createdAt.Month(), createdAt.Day(),
			createdAt.Hour(), createdAt.Minute(), createdAt.Second(), createdAt.Nanosecond(),
			time.UTC,
		)

		result = append(result, r)
	}

	return containersTotal, result, mapSQLErrors(rows.Err())
}

func (r *repository) DeleteContainer(ctx context.Context, namespace, name string) error {
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

	_, err = deleteQuery(ctx, r.db, psql.
		Delete("containers").
		Where(sq.Eq{
			"name":         name,
			"namespace_id": namespaceID,
		}))
	if err != nil {
		return mapSQLErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return mapSQLErrors(err)
	}
	return nil
}
