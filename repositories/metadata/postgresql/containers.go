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
		).
		Values(
			name,
		))
	return mapSQLErrors(err)
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
