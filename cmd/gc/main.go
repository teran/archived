package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/teran/archived/gc/service"
	"github.com/teran/archived/repositories/metadata/postgresql"
)

var (
	appVersion     = "n/a (dev build)"
	buildTimestamp = "undefined"
)

type config struct {
	LogLevel log.Level `envconfig:"LOG_LEVEL" default:"info"`

	MetadataDSN string `envconfig:"METADATA_DSN" required:"true"`

	DryRun                   bool          `envconfig:"DRY_RUN" default:"true"`
	UnpublishedVersionMaxAge time.Duration `envconfig:"UNPUBLISHED_VERSION_MAX_AGE" default:"168h"`
}

func main() {
	var cfg config
	envconfig.MustProcess("", &cfg)

	log.SetLevel(cfg.LogLevel)

	lf := new(log.TextFormatter)
	lf.FullTimestamp = true
	log.SetFormatter(lf)

	log.Infof("Initializing archived-gc (%s @ %s) ...", appVersion, buildTimestamp)

	db, err := sql.Open("postgres", cfg.MetadataDSN)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	postgresqlRepo := postgresql.New(db)

	svc, err := service.New(&service.Config{
		MdRepo:                   postgresqlRepo,
		DryRun:                   cfg.DryRun,
		UnpublishedVersionMaxAge: cfg.UnpublishedVersionMaxAge,
		TimeNowFunc:              time.Now,
	})
	if err != nil {
		panic(err)
	}

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return svc.Run(ctx)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
