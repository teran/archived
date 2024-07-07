package main

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	htmlPresenter "github.com/teran/archived/presenter/access/html"
	"github.com/teran/archived/repositories/metadata/postgresql"
	"github.com/teran/archived/service"
)

var (
	appVersion     = "n/a (dev build)"
	buildTimestamp = "undefined"
)

type config struct {
	Addr        string `envconfig:"ADDR" default:":8080"`
	MetricsAddr string `envconfig:"METRICS_ADDR" default:":8081"`

	LogLevel log.Level `envconfig:"LOG_LEVEL" default:"info"`

	MetadataDSN string `envconfig:"METADATA_DSN" required:"true"`

	HTMLTemplateDir string `envconfig:"HTML_TEMPLATE_DIR" required:"true"`
}

func main() {
	var cfg config
	envconfig.MustProcess("", &cfg)

	log.SetLevel(cfg.LogLevel)

	lf := new(log.TextFormatter)
	lf.FullTimestamp = true
	log.SetFormatter(lf)

	log.Infof("Initializing archived-access (%s @ %s) ...", appVersion, buildTimestamp)

	g, _ := errgroup.WithContext(context.Background())

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	db, err := sql.Open("postgres", cfg.MetadataDSN)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	postgresqlRepo := postgresql.New(db)

	// FIXME: Add BLOB repo initialization
	AccessSvc := service.NewAccessService(postgresqlRepo, nil)

	p := htmlPresenter.New(AccessSvc, cfg.HTMLTemplateDir)
	p.Register(e)

	g.Go(func() error {
		srv := &http.Server{
			Addr:    cfg.Addr,
			Handler: e,
		}

		return srv.ListenAndServe()
	})

	g.Go(func() error {
		http.Handle("/metrics", promhttp.Handler())
		return http.ListenAndServe(cfg.MetricsAddr, nil)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
