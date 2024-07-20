package postgresql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

func (r *repository) CreateContainer(ctx context.Context, name string) error {
	_, err := psql.
		Insert("containers").
		Columns(
			"name",
		).
		Values(
			name,
		).
		RunWith(r.db).
		ExecContext(ctx)

	return errors.Wrap(err, "error executing SQL query")
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
	_, err := psql.
		Delete("containers").
		Where(sq.Eq{"name": name}).
		RunWith(r.db).
		ExecContext(ctx)

	return errors.Wrap(err, "error executing SQL query")
}
