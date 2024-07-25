package main

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/teran/archived/exporter/service"
	"github.com/teran/archived/repositories/metadata/postgresql"
)

var (
	appVersion     = "n/a (dev build)"
	buildTimestamp = "undefined"
)

type config struct {
	Addr string `envconfig:"METRICS_ADDR" default:":8081"`

	LogLevel log.Level `envconfig:"LOG_LEVEL" default:"info"`

	MetadataDSN string `envconfig:"METADATA_DSN" required:"true"`
}

func main() {
	var cfg config
	envconfig.MustProcess("", &cfg)

	log.SetLevel(cfg.LogLevel)

	lf := new(log.TextFormatter)
	lf.FullTimestamp = true
	log.SetFormatter(lf)

	log.Infof("Initializing archived-exporter (%s @ %s) ...", appVersion, buildTimestamp)

	db, err := sql.Open("postgres", cfg.MetadataDSN)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	postgresqlRepo := postgresql.New(db)

	svc, err := service.New(postgresqlRepo)
	if err != nil {
		panic(err)
	}

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return svc.Run(ctx)
	})

	g.Go(func() error {
		http.Handle("/metrics", promhttp.Handler())
		return http.ListenAndServe(cfg.Addr, nil)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
