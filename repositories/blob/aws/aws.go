package aws

import (
	"context"
	"net/url"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"

	"github.com/teran/archived/repositories/blob"
)

var _ blob.Repository = (*s3driver)(nil)

type s3driver struct {
	cli     *s3.Client
	presign *s3.PresignClient
	bucket  string
	ttl     time.Duration
}

func New(cli *s3.Client, bucket string, ttl time.Duration) blob.Repository {
	return &s3driver{
		cli:     cli,
		presign: s3.NewPresignClient(cli),
		bucket:  bucket,
		ttl:     ttl,
	}
}

func (s *s3driver) PutBlobURL(ctx context.Context, key string) (string, error) {
	result, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(s.ttl))

	if err != nil {
		return "", errors.Wrap(err, "error signing URL")
	}
	return result.URL, nil
}

func (s *s3driver) GetBlobURL(ctx context.Context, key, mimeType, filename string) (string, error) {
	disposition := `attachment; filename="` + url.QueryEscape(path.Base(filename)) + `"`

	result, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.bucket),
		Key:                        aws.String(key),
		ResponseContentType:        aws.String(mimeType),
		ResponseContentDisposition: aws.String(disposition),
	}, s3.WithPresignExpires(s.ttl))

	if err != nil {
		return "", errors.Wrap(err, "error signing URL")
	}
	return result.URL, nil
}
