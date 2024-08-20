package postgresql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
)

func (r *repository) CreateContainer(ctx context.Context, name string) error {
	_, err := insertQuery(ctx, r.db, psql.
		Insert("containers").
		Columns(
			"name",
			"created_at",
		).
		Values(
			name,
			r.tp().UTC(),
		))
	return mapSQLErrors(err)
}

func (r *repository) RenameContainer(ctx context.Context, oldName, newName string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return mapSQLErrors(err)
	}
	defer tx.Rollback()

	row, err := selectQueryRow(ctx, tx, psql.
		Select("id").
		From("containers").
		Where(sq.Eq{"name": oldName}))
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

func (r *repository) ListContainers(ctx context.Context) ([]string, error) {
	rows, err := selectQuery(ctx, r.db, psql.
		Select("name").
		From("containers").
		OrderBy("name"))
	if err != nil {
		return nil, mapSQLErrors(err)
	}
	defer rows.Close()

	result := []string{}
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return nil, mapSQLErrors(err)
		}

		result = append(result, r)
	}

	return result, nil
}

func (r *repository) DeleteContainer(ctx context.Context, name string) error {
	_, err := deleteQuery(ctx, r.db, psql.
		Delete("containers").
		Where(sq.Eq{"name": name}))
	return mapSQLErrors(err)
}
