package postgresql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	ptr "github.com/teran/go-ptr"
)

const defaultLimit uint64 = 1000

func (r *repository) CreateVersion(ctx context.Context, container string) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", errors.Wrap(err, "error beginning transaction")
	}
	defer tx.Rollback()

	row := psql.
		Select("id").
		From("containers").
		Where(sq.Eq{"name": container}).
		RunWith(tx).
		QueryRowContext(ctx)

	var containerID uint
	if err := row.Scan(&containerID); err != nil {
		return "", errors.Wrap(err, "error looking up container")
	}

	versionID := r.tp().UTC().Format("20060102150405")

	_, err = psql.
		Insert("versions").
		Columns(
			"container_id",
			"name",
			"is_published",
		).
		Values(
			containerID,
			versionID,
			false,
		).
		RunWith(tx).
		ExecContext(ctx)
	if err != nil {
		return "", errors.Wrap(err, "error executing SQL query")
	}

	if err := tx.Commit(); err != nil {
		return "", errors.Wrap(err, "error committing transaction")
	}

	return versionID, nil
}

func (r *repository) ListPublishedVersionsByContainer(ctx context.Context, container string) ([]string, error) {
	_, versions, err := r.listVersionsByContainer(ctx, container, ptr.Bool(true), 0, 0)
	return versions, err
}

func (r *repository) ListAllVersionsByContainer(ctx context.Context, container string) ([]string, error) {
	_, versions, err := r.listVersionsByContainer(ctx, container, nil, 0, 0)
	return versions, err
}

func (r *repository) ListPublishedVersionsByContainerAndPage(ctx context.Context, container string, offset, limit uint64) (uint64, []string, error) {
	return r.listVersionsByContainer(ctx, container, nil, offset, limit)
}

func (r *repository) listVersionsByContainer(ctx context.Context, container string, isPublished *bool, offset, limit uint64) (uint64, []string, error) {
	if limit == 0 {
		limit = defaultLimit
	}

	row := psql.
		Select("id").
		From("containers").
		Where(sq.Eq{"name": container}).
		RunWith(r.db).
		QueryRowContext(ctx)

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

	row = psql.
		Select("COUNT(*)").
		From("versions").
		RunWith(r.db).
		QueryRowContext(ctx)

	var versionsTotal uint64
	if err := row.Scan(&versionsTotal); err != nil {
		return 0, nil, mapSQLErrors(err)
	}

	rows, err := psql.
		Select("name").
		From("versions").
		Where(condition).
		OrderBy("id").
		Offset(offset).
		Limit(limit).
		RunWith(r.db).
		QueryContext(ctx)
	if err != nil {
		return 0, nil, errors.Wrap(err, "error executing SQL query")
	}
	defer rows.Close()

	result := []string{}
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return 0, nil, errors.Wrap(err, "error decoding database result")
		}

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
	row := psql.
		Select("id").
		From("containers").
		Where(sq.Eq{"name": container}).
		RunWith(tx).
		QueryRowContext(ctx)

	if err := row.Scan(&containerID); err != nil {
		return errors.Wrap(err, "error looking up container")
	}

	_, err = psql.
		Update("versions").
		Set("is_published", true).
		Where(sq.Eq{
			"container_id": containerID,
			"name":         version,
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
