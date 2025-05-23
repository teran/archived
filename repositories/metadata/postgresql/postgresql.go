package postgresql

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/repositories/metadata"
)

const defaultNamespace = "default"

var (
	psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	queryCountTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "archived",
		Subsystem: "repository",
		Name:      "query_count_total",
		Help:      "Total time of SQL queries by kind",
	}, []string{"kind"})

	queryTimeTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "archived",
		Subsystem: "repository",
		Name:      "query_time_seconds_total",
		Help:      "Total time of SQL queries by kind",
	}, []string{"kind"})
)

type repository struct {
	db *sql.DB
	tp func() time.Time
}

func init() {
	prometheus.MustRegister(queryCountTotal)
	prometheus.MustRegister(queryTimeTotal)
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
	if errors.Is(err, sql.ErrNoRows) {
		return metadata.ErrNotFound
	}

	if err, ok := err.(*pq.Error); ok { //nolint:errorlint
		if err.Code == "23505" {
			return metadata.ErrConflict
		}
	}

	return err
}

type queryRunner interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type execRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func selectQueryRow(ctx context.Context, db queryRunner, q sq.SelectBuilder) (sq.RowScanner, error) {
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	start := time.Now()
	defer func() {
		since := time.Since(start)

		queryCountTotal.WithLabelValues("select").Inc()
		queryTimeTotal.WithLabelValues("select").Add(since.Seconds())

		log.WithFields(log.Fields{
			"query":    sql,
			"args":     args,
			"duration": since,
		}).Debug("SQL query executed")
	}()

	return db.QueryRowContext(ctx, sql, args...), nil
}

func selectQuery(ctx context.Context, db queryRunner, q sq.SelectBuilder) (*sql.Rows, error) {
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	start := time.Now()
	defer func() {
		since := time.Since(start)

		queryCountTotal.WithLabelValues("select").Inc()
		queryTimeTotal.WithLabelValues("select").Add(since.Seconds())

		log.WithFields(log.Fields{
			"query":    sql,
			"args":     args,
			"duration": since,
		}).Debug("SQL query executed")
	}()

	return db.QueryContext(ctx, sql, args...) //nolint:sqlclosecheck
}

func insertQuery(ctx context.Context, db execRunner, q sq.InsertBuilder) (sql.Result, error) { //nolint:unparam
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	start := time.Now()
	defer func() {
		since := time.Since(start)

		queryCountTotal.WithLabelValues("insert").Inc()
		queryTimeTotal.WithLabelValues("insert").Add(since.Seconds())

		log.WithFields(log.Fields{
			"query":    sql,
			"args":     args,
			"duration": since,
		}).Debug("SQL query executed")
	}()

	return db.ExecContext(ctx, sql, args...)
}

func insertQueryRow(ctx context.Context, db queryRunner, q sq.InsertBuilder) (sq.RowScanner, error) {
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	start := time.Now()
	defer func() {
		since := time.Since(start)

		queryCountTotal.WithLabelValues("insert").Inc()
		queryTimeTotal.WithLabelValues("insert").Add(since.Seconds())

		log.WithFields(log.Fields{
			"query":    sql,
			"args":     args,
			"duration": since,
		}).Debug("SQL query executed")
	}()

	return db.QueryRowContext(ctx, sql, args...), nil
}

func updateQuery(ctx context.Context, db execRunner, q sq.UpdateBuilder) (sql.Result, error) { //nolint:unparam
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	start := time.Now()
	defer func() {
		since := time.Since(start)

		queryCountTotal.WithLabelValues("update").Inc()
		queryTimeTotal.WithLabelValues("update").Add(since.Seconds())

		log.WithFields(log.Fields{
			"query":    sql,
			"args":     args,
			"duration": since,
		}).Debug("SQL query executed")
	}()

	return db.ExecContext(ctx, sql, args...)
}

func deleteQuery(ctx context.Context, db execRunner, q sq.DeleteBuilder) (sql.Result, error) { //nolint:unparam
	sql, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	start := time.Now()
	defer func() {
		since := time.Since(start)

		queryCountTotal.WithLabelValues("delete").Inc()
		queryTimeTotal.WithLabelValues("delete").Add(since.Seconds())

		log.WithFields(log.Fields{
			"query":    sql,
			"args":     args,
			"duration": since,
		}).Debug("SQL query executed")
	}()

	return db.ExecContext(ctx, sql, args...)
}
