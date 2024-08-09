package postgresql

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	ptr "github.com/teran/go-ptr"

	"github.com/teran/archived/models"
)

const defaultLimit uint64 = 1000

func (r *repository) CreateVersion(ctx context.Context, container string) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", errors.Wrap(err, "error beginning transaction")
	}
	defer tx.Rollback()

	row, err := selectQueryRow(ctx, tx, psql.
		Select("id").
		From("containers").
		Where(sq.Eq{"name": container}))
	if err != nil {
		return "", mapSQLErrors(err)
	}

	var containerID uint
	if err := row.Scan(&containerID); err != nil {
		return "", errors.Wrap(err, "error looking up container")
	}

	versionID := r.tp().UTC().Format("20060102150405")

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
			r.tp().UTC(),
		))
	if err != nil {
		return "", mapSQLErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return "", errors.Wrap(err, "error committing transaction")
	}

	return versionID, nil
}

func (r *repository) GetLatestPublishedVersionByContainer(ctx context.Context, container string) (string, error) {
	row, err := selectQueryRow(ctx, r.db, psql.
		Select("v.name").
		From("versions v").
		Join("containers c ON v.container_id = c.id").
		Where(sq.Eq{
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

func (r *repository) ListPublishedVersionsByContainer(ctx context.Context, container string) ([]models.Version, error) {
	_, versions, err := r.listVersionsByContainer(ctx, container, ptr.Bool(true), 0, 0)
	return versions, err
}

func (r *repository) ListAllVersionsByContainer(ctx context.Context, container string) ([]models.Version, error) {
	_, versions, err := r.listVersionsByContainer(ctx, container, nil, 0, 0)
	return versions, err
}

func (r *repository) ListUnpublishedVersionsByContainer(ctx context.Context, container string) ([]models.Version, error) {
	_, versions, err := r.listVersionsByContainer(ctx, container, ptr.Bool(false), 0, 0)
	return versions, err
}

func (r *repository) ListPublishedVersionsByContainerAndPage(ctx context.Context, container string, offset, limit uint64) (uint64, []models.Version, error) {
	return r.listVersionsByContainer(ctx, container, ptr.Bool(true), offset, limit)
}

func (r *repository) listVersionsByContainer(ctx context.Context, container string, isPublished *bool, offset, limit uint64) (uint64, []models.Version, error) {
	if limit == 0 {
		limit = defaultLimit
	}

	row, err := selectQueryRow(ctx, r.db, psql.
		Select("id").
		From("containers").
		Where(sq.Eq{"name": container}))
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
			return 0, nil, errors.Wrap(err, "error decoding database result")
		}
		r.CreatedAt = time.Date(
			createdAt.Year(), createdAt.Month(), createdAt.Day(),
			createdAt.Hour(), createdAt.Minute(), createdAt.Second(), createdAt.Nanosecond(),
			time.UTC,
		)

		result = append(result, r)
	}

	return versionsTotal, result, nil
}

func (r *repository) MarkVersionPublished(ctx context.Context, container, version string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "error beginning transaction")
	}
	defer tx.Rollback()

	var containerID uint

	row, err := selectQueryRow(ctx, tx, psql.
		Select("id").
		From("containers").
		Where(sq.Eq{"name": container}))
	if err != nil {
		return mapSQLErrors(err)
	}

	if err := row.Scan(&containerID); err != nil {
		return errors.Wrap(err, "error looking up container")
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
		return errors.Wrap(err, "error committing transaction")
	}

	return nil
}

func (r *repository) DeleteVersion(ctx context.Context, container, version string) error {
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

	var versionID uint64
	if err := row.Scan(&versionID); err != nil {
		return errors.Wrap(err, "error looking up version")
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
		return errors.Wrap(err, "error committing transaction")
	}
	return nil
}
