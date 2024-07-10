package main

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	grpcManagePresenter "github.com/teran/archived/presenter/manage/grpc"
)

var (
	appVersion     = "n/a (dev build)"
	buildTimestamp = "undefined"
)

type config struct {
	Addr        string `envconfig:"ADDR" default:":5555"`
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
}

func main() {
	var cfg config
	envconfig.MustProcess("", &cfg)

	log.SetLevel(cfg.LogLevel)

	lf := new(log.TextFormatter)
	lf.FullTimestamp = true
	log.SetFormatter(lf)

	log.Infof("Initializing archived-manage (%s @ %s) ...", appVersion, buildTimestamp)

	g, _ := errgroup.WithContext(context.Background())

	managePresenter := grpcManagePresenter.New()

	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		panic(err)
	}

	gs := grpc.NewServer()
	managePresenter.Register(gs)

	g.Go(func() error {
		return gs.Serve(listener)
	})

	g.Go(func() error {
		http.Handle("/metrics", promhttp.Handler())
		return http.ListenAndServe(cfg.MetricsAddr, nil)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}
