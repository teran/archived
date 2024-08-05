package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpcManagePresenter "github.com/teran/archived/presenter/manager/grpc"
	awsBlobRepo "github.com/teran/archived/repositories/blob/aws"
	"github.com/teran/archived/repositories/metadata/postgresql"
	"github.com/teran/archived/service"
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
	BLOBS3CreateBucket     bool          `envconfig:"BLOB_S3_CREATE_BUCKET" default:"false"`
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

	log.Infof("Initializing archived-manager (%s @ %s) ...", appVersion, buildTimestamp)

	g, _ := errgroup.WithContext(context.Background())

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

	s3client := s3.New(awsSession)

	if cfg.BLOBS3CreateBucket {
		_, err := s3client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(cfg.BLOBS3Bucket),
		})
		if err != nil {
			panic(err)
		}
	}
	blobRepo := awsBlobRepo.New(s3client, cfg.BLOBS3Bucket, cfg.BLOBS3PresignedLinkTTL)

	managerSvc := service.NewManager(postgresqlRepo, blobRepo)

	managePresenter := grpcManagePresenter.New(managerSvc)

	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		panic(err)
	}

	gs := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(interceptorLogger()),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(interceptorLogger()),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		),
	)
	managePresenter.Register(gs)

	g.Go(func() error {
		return gs.Serve(listener)
	})

	g.Go(func() error {
		http.Handle("/metrics", promhttp.Handler())

		http.HandleFunc("/healthz/startup", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok\n"))
		})

		http.HandleFunc("/healthz/readiness", func(w http.ResponseWriter, r *http.Request) {
			if err := db.Ping(); err != nil {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok\n"))
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("failed\n"))
			}
		})

		http.HandleFunc("/healthz/liveness", func(w http.ResponseWriter, r *http.Request) {
			if err := db.Ping(); err != nil {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok\n"))
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("failed\n"))
			}
		})

		return http.ListenAndServe(cfg.MetricsAddr, nil)
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}
}

func grpcPanicRecoveryHandler(p any) (err error) {
	log.Errorf("recovered from panic: %s", debug.Stack())
	return status.Errorf(codes.Internal, "%s", p)
}

func interceptorLogger() logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			log.Debugf("%s: %s", msg, formatFields(fields))
		case logging.LevelInfo:
			log.Infof("%s: %s", msg, formatFields(fields))
		case logging.LevelWarn:
			log.Warnf("%s: %s", msg, formatFields(fields))
		case logging.LevelError:
			log.Errorf("%s: %s", msg, formatFields(fields))
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func formatFields(in []any) (out string) {
	for i := 1; i < len(in); i += 2 {
		key := in[i-1]
		value := in[i]

		out += fmt.Sprintf("%v=%v; ", key, value)
	}
	return out
}
