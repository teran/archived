package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	awsBlobRepo "github.com/teran/archived/repositories/blob/aws"
	"github.com/teran/archived/repositories/metadata/postgresql"
	"github.com/teran/archived/service"
)

var (
	appVersion     = "n/a (dev build)"
	buildTimestamp = "undefined"
)

type config struct {
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

	CreateNamespaces             int `envconfig:"CREATE_NAMESPACES" default:"10"`
	CreateContainersPerNamespace int `envconfig:"CREATE_CONTAINERS_PER_NAMESPACE" default:"100"`
	CreateVersionsPerContainer   int `envconfig:"CREATE_VERSIONS_PER_CONTAINER" default:"100"`
	CreateObjectsPerVersion      int `envconfig:"CREATE_OBJECTS_PER_VERSION" default:"100"`
	MaxObjectSizeBytes           int `envconfig:"MAX_OBJECT_SIZE_BYTES" default:"4096"`
}

func main() {
	var cfg config
	envconfig.MustProcess("", &cfg)

	log.SetLevel(cfg.LogLevel)

	lf := new(log.TextFormatter)
	lf.FullTimestamp = true
	log.SetFormatter(lf)

	log.Infof("Initializing archived-seeder (%s @ %s) ...", appVersion, buildTimestamp)

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
	if err != nil {
		panic(err)
	}

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

	ctx := context.TODO()

	for i := 0; i <= cfg.CreateNamespaces; i++ {
		namespace := fmt.Sprintf("namespace-%06d", i)
		if err := managerSvc.CreateNamespace(ctx, namespace); err != nil {
			panic(err)
		}
		for j := 0; j <= cfg.CreateContainersPerNamespace; j++ {
			container := fmt.Sprintf("container-%06d", j)
			if err := managerSvc.CreateContainer(ctx, namespace, container, -1); err != nil {
				panic(err)
			}
			for k := 0; k <= cfg.CreateVersionsPerContainer; k++ {
				version, err := managerSvc.CreateVersion(ctx, namespace, container)
				if err != nil {
					panic(err)
				}
				for l := 0; l <= cfg.CreateObjectsPerVersion; l++ {
					key := fmt.Sprintf("object-%06d", l)

					size, err := rand.Int(rand.Reader, big.NewInt(int64(cfg.MaxObjectSizeBytes)))
					if err != nil {
						panic(err)
					}

					data := make([]byte, size.Int64())
					if _, err := rand.Read(data); err != nil {
						panic(err)
					}

					h := sha256.New()
					if _, err := io.Copy(h, bytes.NewReader(data)); err != nil {
						panic(err)
					}
					casKey := hex.EncodeToString(h.Sum(nil))

					url, err := managerSvc.EnsureBLOBPresenceOrGetUploadURL(ctx, casKey, uint64(len(data)), "application/octet-stream")
					if err != nil {
						panic(err)
					}

					req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
					if err != nil {
						panic(err)
					}

					req.Header.Set("Content-Type", "multipart/form-data")
					uploadResp, err := http.DefaultClient.Do(req)
					if err != nil {
						panic(err)
					}
					defer func() { _ = uploadResp.Body.Close() }()

					if uploadResp.StatusCode > 299 {
						panic(errors.Errorf("unexpected status code on upload: %s", uploadResp.Status))
					}

					if err := managerSvc.AddObject(ctx, namespace, container, version, key, casKey); err != nil {
						panic(err)
					}
				}

				if err := managerSvc.PublishVersion(ctx, namespace, container, version); err != nil {
					panic(err)
				}
			}
		}
	}
}
