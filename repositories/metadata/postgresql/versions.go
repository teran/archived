package postgresql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	ptr "github.com/teran/go-ptr"
)

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
	return r.listVersionsByContainer(ctx, container, ptr.Bool(true))
}

func (r *repository) ListAllVersionsByContainer(ctx context.Context, container string) ([]string, error) {
	return r.listVersionsByContainer(ctx, container, nil)
}

func (r *repository) listVersionsByContainer(ctx context.Context, container string, isPublished *bool) ([]string, error) {
	row := psql.
		Select("id").
		From("containers").
		Where(sq.Eq{"name": container}).
		RunWith(r.db).
		QueryRowContext(ctx)

	var containerID uint
	if err := row.Scan(&containerID); err != nil {
		return nil, errors.Wrap(err, "error looking up container")
	}

	condition := sq.Eq{
		"container_id": containerID,
	}

	if isPublished != nil {
		condition["is_published"] = *isPublished
	}

	rows, err := psql.
		Select("name").
		From("versions").
		Where(condition).
		OrderBy("id").
		RunWith(r.db).
		QueryContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error executing SQL query")
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
	if err != nil {
		return errors.Wrap(err, "error generating SQL query")
	}

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
