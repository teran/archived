package aws

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	"github.com/teran/archived/repositories/blob"
	"github.com/teran/go-docker-testsuite/applications/minio"
)

func (s *repoTestSuite) TestAll() {
	url, err := s.driver.PutBlobURL(s.ctx, "blah/test/key.txt")
	s.Require().NoError(err)

	err = uploadToURL(s.ctx, url, []byte("test data"))
	s.Require().NoError(err)

	url, err = s.driver.GetBlobURL(s.ctx, "blah/test/key.txt", "application/json", "test-file.txt")
	s.Require().NoError(err)

	data, mimeType, disposition, err := fetchURL(s.ctx, url)
	s.Require().NoError(err)
	s.Require().Equal("application/json", mimeType)
	s.Require().Equal(`attachment; filename="test-file.txt"`, disposition)
	s.Require().Equal("test data", string(data))
}

// Definitions ...
type repoTestSuite struct {
	suite.Suite

	ctx    context.Context
	cli    *s3.Client
	driver blob.Repository
	minio  minio.Minio
}

func (s *repoTestSuite) SetupSuite() {
	s.ctx = context.Background()

	app, err := minio.New(s.ctx)
	s.Require().NoError(err)

	s.minio = app
}

func (s *repoTestSuite) SetupTest() {
	endpoint, err := s.minio.GetEndpointURL()
	s.Require().NoError(err)

	cfg, err := config.LoadDefaultConfig(s.T().Context(),
		config.WithRegion("default"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				minio.MinioAccessKey,
				minio.MinioAccessKeySecret,
				"",
			),
		),
	)
	s.Require().NoError(err)

	s.cli = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("http://" + endpoint)
		o.UsePathStyle = true
		o.EndpointOptions.DisableHTTPS = true
	})

	_, err = s.cli.CreateBucket(s.T().Context(), &s3.CreateBucketInput{
		Bucket: aws.String("test-bucket"),
	})
	s.Require().NoError(err)

	s.driver = New(s.cli, "test-bucket", 5*time.Second)
}

func (s *repoTestSuite) TearDownTest() {
	out, err := s.cli.ListObjects(s.T().Context(), &s3.ListObjectsInput{
		Bucket: aws.String("test-bucket"),
	})
	s.Require().NoError(err)

	for _, obj := range out.Contents {
		_, err := s.cli.DeleteObject(s.T().Context(), &s3.DeleteObjectInput{
			Bucket: aws.String("test-bucket"),
			Key:    obj.Key,
		})
		s.Require().NoError(err)
	}

	_, err = s.cli.DeleteBucket(s.T().Context(), &s3.DeleteBucketInput{
		Bucket: aws.String("test-bucket"),
	})
	s.Require().NoError(err)
}

func (s *repoTestSuite) TearDownSuite() {
	err := s.minio.Close(s.ctx)
	s.Require().NoError(err)
}

func TestRepoTestSuite(t *testing.T) {
	suite.Run(t, &repoTestSuite{})
}

func fetchURL(ctx context.Context, url string) ([]byte, string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", "", err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	return data, resp.Header.Get("Content-Type"), resp.Header.Get("Content-Disposition"), err
}

func uploadToURL(ctx context.Context, url string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "multipart/form-data")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("status code %d != %d", http.StatusOK, resp.StatusCode)
	}
	return nil
}
