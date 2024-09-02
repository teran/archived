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

		http.HandleFunc("/healthz/startup", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("ok\n")); err != nil {
				panic(err)
			}
		})

		http.HandleFunc("/healthz/readiness", func(w http.ResponseWriter, r *http.Request) {
			if err := db.Ping(); err != nil {
				log.Warnf("db.Ping() error on readiness probe: %s", err)

				w.WriteHeader(http.StatusServiceUnavailable)
				if _, err := w.Write([]byte("failed\n")); err != nil {
					panic(err)
				}
			} else {
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte("ok\n")); err != nil {
					panic(err)
				}
			}
		})

		http.HandleFunc("/healthz/liveness", func(w http.ResponseWriter, r *http.Request) {
			if err := db.Ping(); err != nil {
				log.Warnf("db.Ping() error on liveness probe: %s", err)

				w.WriteHeader(http.StatusServiceUnavailable)
				if _, err := w.Write([]byte("failed\n")); err != nil {
					panic(err)
				}
			} else {
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte("ok\n")); err != nil {
					panic(err)
				}
			}
		})

		return http.ListenAndServe(cfg.Addr, nil)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
