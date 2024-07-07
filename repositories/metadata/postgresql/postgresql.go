package postgresql

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"

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
