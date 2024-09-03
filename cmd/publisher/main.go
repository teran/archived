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
	memcacheCli "github.com/bradfitz/gomemcache/memcache"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo-contrib/echoprometheus"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/teran/archived/appmetrics"
	htmlPresenter "github.com/teran/archived/publisher/presenter/html"
	awsBlobRepo "github.com/teran/archived/repositories/blob/aws"
	"github.com/teran/archived/repositories/cache/metadata/memcache"
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

	MemcacheServers []string      `envconfig:"MEMCACHE_SERVERS"`
	MemcacheTTL     time.Duration `envconfig:"MEMCACHE_TTL" default:"60m"`

	BLOBS3Endpoint         string        `envconfig:"BLOB_S3_ENDPOINT" required:"true"`
	BLOBS3Bucket           string        `envconfig:"BLOB_S3_BUCKET" required:"true"`
	BLOBS3PresignedLinkTTL time.Duration `envconfig:"BLOB_S3_PRESIGNED_LINK_TTL" default:"5m"`
	BLOBS3AccessKeyID      string        `envconfig:"BLOB_S3_ACCESS_KEY_ID" required:"true"`
	BLOBS3SecretKey        string        `envconfig:"BLOB_S3_SECRET_KEY" required:"true"`
	BLOBS3Region           string        `envconfig:"BLOB_S3_REGION" default:"default"`
	BLOBS3DisableSSL       bool          `envconfig:"BLOB_S3_DISABLE_SSL" default:"false"`
	BLOBS3ForcePathStyle   bool          `envconfig:"BLOB_S3_FORCE_PATH_STYLE" default:"true"`

	BLOBS3PreserveSchemeOnRedirect bool `envconfig:"BLOB_S3_PRESERVE_SCHEME_ON_REDIRECT" default:"true"`

	HTMLTemplateDir string `envconfig:"HTML_TEMPLATE_DIR" required:"true"`
	StaticDir       string `envconfig:"STATIC_DIR" required:"true"`

	VersionsPerPage   uint64 `envconfig:"VERSIONS_PER_PAGE" default:"50"`
	ObjectsPerPage    uint64 `envconfig:"OBJECTS_PER_PAGE" default:"50"`
	ContainersPerPage uint64 `envconfig:"CONTAINERS_PER_PAGE" default:"50"`
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

	repo := postgresql.New(db)

	var cli *memcacheCli.Client
	if len(cfg.MemcacheServers) > 0 {
		log.Debugf(
			"%d memcache servers specified for metadata caching. Initializing read-through cache ...",
			len(cfg.MemcacheServers),
		)

		cli = memcacheCli.New(cfg.MemcacheServers...)
		if err := cli.Ping(); err != nil {
			panic(err)
		}

		repo = memcache.New(cli, repo, cfg.MemcacheTTL, "publisher")
	}

	awsSession, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(cfg.BLOBS3Endpoint),
		Region:           aws.String(cfg.BLOBS3Region),
		DisableSSL:       aws.Bool(cfg.BLOBS3DisableSSL),
		S3ForcePathStyle: aws.Bool(cfg.BLOBS3ForcePathStyle),
		Credentials: credentials.NewStaticCredentials(
			cfg.BLOBS3AccessKeyID, cfg.BLOBS3SecretKey, "",
		),
	})
	if err != nil {
		panic(err)
	}

	blobRepo := awsBlobRepo.New(s3.New(awsSession), cfg.BLOBS3Bucket, cfg.BLOBS3PresignedLinkTTL)

	publisherSvc := service.NewPublisher(repo, blobRepo, cfg.VersionsPerPage, cfg.ObjectsPerPage, cfg.ContainersPerPage)

	p := htmlPresenter.New(publisherSvc, cfg.HTMLTemplateDir, cfg.StaticDir, cfg.BLOBS3PreserveSchemeOnRedirect)
	p.Register(e)

	g.Go(func() error {
		srv := &http.Server{
			Addr:    cfg.Addr,
			Handler: e,
		}

		return srv.ListenAndServe()
	})

	me := echo.New()
	me.Use(middleware.Logger())
	me.Use(echoprometheus.NewMiddleware("publisher_metrics"))
	me.Use(middleware.Recover())

	checkFn := func() error {
		if len(cfg.MemcacheServers) > 0 {
			if err := cli.Ping(); err != nil {
				return err
			}
		}

		if err := db.Ping(); err != nil {
			return err
		}

		return nil
	}

	metrics := appmetrics.New(checkFn, checkFn, checkFn)
	metrics.Register(me)

	g.Go(func() error {
		srv := http.Server{
			Addr:    cfg.MetricsAddr,
			Handler: me,
		}

		return srv.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
