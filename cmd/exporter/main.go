package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo-contrib/echoprometheus"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/teran/appmetrics"
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

	ObserveInterval time.Duration `envconfig:"OBSERVE_INTERVAL" default:"60s"`
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

	svc, err := service.New(postgresqlRepo, cfg.ObserveInterval)
	if err != nil {
		panic(err)
	}

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return svc.Run(ctx)
	})

	me := echo.New()
	me.Use(middleware.Logger())
	me.Use(echoprometheus.NewMiddleware("exporter_metrics"))
	me.Use(middleware.Recover())

	checkFn := func() error {
		if err := db.Ping(); err != nil {
			return err
		}

		return nil
	}

	metrics := appmetrics.New(checkFn, checkFn, checkFn)
	metrics.Register(me)

	g.Go(func() error {
		srv := http.Server{
			Addr:    cfg.Addr,
			Handler: me,
		}

		return srv.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
