package postgresql

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	log "github.com/sirupsen/logrus"
)

func (r *repository) CreateNamespace(ctx context.Context, name string) error {
	_, err := insertQuery(ctx, r.db, psql.
		Insert("namespaces").
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

func (r *repository) RenameNamespace(ctx context.Context, oldName, newName string) error {
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
		Where(sq.Eq{"name": oldName}))
	if err != nil {
		return mapSQLErrors(err)
	}

	var namespaceID uint
	if err := row.Scan(&namespaceID); err != nil {
		return mapSQLErrors(err)
	}

	_, err = updateQuery(ctx, tx, psql.
		Update("namespaces").
		Set("name", newName).
		Where(sq.Eq{
			"id": namespaceID,
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

func (r *repository) ListNamespaces(ctx context.Context) ([]string, error) {
	rows, err := selectQuery(ctx, r.db, psql.
		Select("name").
		From("namespaces").
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

func (r *repository) DeleteNamespace(ctx context.Context, name string) error {
	_, err := deleteQuery(ctx, r.db, psql.
		Delete("namespaces").
		Where(sq.Eq{"name": name}))
	return mapSQLErrors(err)
}
