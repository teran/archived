package postgresql

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/repositories/metadata"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type repository struct {
	db *sql.DB
	tp func() time.Time
}

func New(db *sql.DB) metadata.Repository {
	return newWithTimeProvider(db, time.Now)
}

func newWithTimeProvider(db *sql.DB, tp func() time.Time) metadata.Repository {
	return &repository{
		db: db,
		tp: tp,
	}
}

func mapSQLErrors(err error) error {
	switch err {
	case sql.ErrNoRows:
		return metadata.ErrNotFound
	default:
		return err
	}
}

type queryRunner interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func selectQueryRow(ctx context.Context, db queryRunner, q sq.SelectBuilder) (sq.RowScanner, error) {
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"query": sql,
		"args":  args,
	}).Tracef("SQL query generated")

	return db.QueryRowContext(ctx, sql, args...), nil
}

func selectQuery(ctx context.Context, db queryRunner, q sq.SelectBuilder) (*sql.Rows, error) {
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"query": sql,
		"args":  args,
	}).Tracef("SQL query generated")

	return db.QueryContext(ctx, sql, args...)
}
