package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo-contrib/echoprometheus"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	htmlPresenter "github.com/teran/archived/presenter/publisher/html"
	awsBlobRepo "github.com/teran/archived/repositories/blob/aws"
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

	BLOBS3Endpoint         string        `envconfig:"BLOB_S3_ENDPOINT" required:"true"`
	BLOBS3Bucket           string        `envconfig:"BLOB_S3_BUCKET" required:"true"`
	BLOBS3PresignedLinkTTL time.Duration `envconfig:"BLOB_S3_PRESIGNED_LINK_TTL" default:"5m"`
	BLOBS3AccessKeyID      string        `envconfig:"BLOB_S3_ACCESS_KEY_ID" required:"true"`
	BLOBS3SecretKey        string        `envconfig:"BLOB_S3_SECRET_KEY" required:"true"`
	BLOBS3Region           string        `envconfig:"BLOB_S3_REGION" default:"default"`
	BLOBS3DisableSSL       bool          `envconfig:"BLOB_S3_DISABLE_SSL" default:"false"`
	BLOBS3ForcePathStyle   bool          `envconfig:"BLOB_S3_FORCE_PATH_STYLE" default:"true"`

	HTMLTemplateDir string `envconfig:"HTML_TEMPLATE_DIR" required:"true"`
	StaticDir       string `envconfig:"STATIC_DIR" required:"true"`

	VersionsPerPage uint64 `envconfig:"VERSIONS_PER_PAGE" default:"50"`
	ObjectsPerPage  uint64 `envconfig:"OBJECTS_PER_PAGE" default:"50"`
}

func main() {
	var cfg config
	envconfig.MustProcess("", &cfg)

	log.SetLevel(cfg.LogLevel)

	lf := new(log.TextFormatter)
	lf.FullTimestamp = true
	log.SetFormatter(lf)

	log.Infof("Initializing archived-publisher (%s @ %s) ...", appVersion, buildTimestamp)

	g, _ := errgroup.WithContext(context.Background())

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(echoprometheus.NewMiddleware("publisher"))
	e.Use(middleware.Recover())

	db, err := sql.Open("postgres", cfg.MetadataDSN)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	postgresqlRepo := postgresql.New(db)

	awsSession, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(cfg.BLOBS3Endpoint),
		Region:           aws.String(cfg.BLOBS3Region),
		DisableSSL:       aws.Bool(cfg.BLOBS3DisableSSL),
		S3ForcePathStyle: aws.Bool(cfg.BLOBS3ForcePathStyle),
		Credentials: credentials.NewStaticCredentials(
			cfg.BLOBS3AccessKeyID, cfg.BLOBS3SecretKey, "",
		),
	})

	blobRepo := awsBlobRepo.New(s3.New(awsSession), cfg.BLOBS3Bucket, cfg.BLOBS3PresignedLinkTTL)

	publisherSvc := service.NewPublisher(postgresqlRepo, blobRepo, cfg.VersionsPerPage, cfg.ObjectsPerPage)

	p := htmlPresenter.New(publisherSvc, cfg.HTMLTemplateDir, cfg.StaticDir)
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

		http.HandleFunc("/healthz/startup", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok\n"))
		})

		http.HandleFunc("/healthz/readiness", func(w http.ResponseWriter, r *http.Request) {
			if err := db.Ping(); err != nil {
				log.Warnf("db.Ping() error on readiness probe: %s", err)

				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("failed\n"))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok\n"))
			}
		})

		http.HandleFunc("/healthz/liveness", func(w http.ResponseWriter, r *http.Request) {
			if err := db.Ping(); err != nil {
				log.Warnf("db.Ping() error on liveness probe: %s", err)

				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("failed\n"))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok\n"))
			}
		})

		return http.ListenAndServe(cfg.MetricsAddr, nil)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
