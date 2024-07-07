package aws

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"

	"github.com/teran/archived/repositories/blob"
)

var _ blob.Repository = (*s3driver)(nil)

type s3driver struct {
	cli    *s3.S3
	bucket string
	ttl    time.Duration
}

func New(cli *s3.S3, bucket string, ttl time.Duration) blob.Repository {
	return &s3driver{
		cli:    cli,
		bucket: bucket,
		ttl:    ttl,
	}
}

func (s *s3driver) PutBlob(ctx context.Context, key string, rd io.ReadSeeker) error {
	_, err := s.cli.PutObject(&s3.PutObjectInput{
		Body:   rd,
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return errors.Wrap(err, "error putting object")
}

func (s *s3driver) GetBlobURL(ctx context.Context, key string) (string, error) {
	req, _ := s.cli.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	url, err := req.Presign(s.ttl)
	if err != nil {
		return "", errors.Wrap(err, "error signing URL")
	}

	return url, nil
}
