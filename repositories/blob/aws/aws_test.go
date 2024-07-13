package aws

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

	url, err = s.driver.GetBlobURL(s.ctx, "blah/test/key.txt")
	s.Require().NoError(err)

	data, err := fetchURL(s.ctx, url)
	s.Require().NoError(err)
	s.Require().Equal("test data", string(data))
}

// Definitions ...
type repoTestSuite struct {
	suite.Suite

	ctx    context.Context
	cli    *s3.S3
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

	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("default"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			minio.MinioAccessKey, minio.MinioAccessKeySecret, "",
		),
	})
	s.Require().NoError(err)

	s.cli = s3.New(sess)

	_, err = s.cli.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String("test-bucket"),
	})
	s.Require().NoError(err)

	s.driver = New(s.cli, "test-bucket", 5*time.Second)
}

func (s *repoTestSuite) TearDownTest() {
	out, err := s.cli.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String("test-bucket"),
	})
	s.Require().NoError(err)

	for _, obj := range out.Contents {
		_, err := s.cli.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String("test-bucket"),
			Key:    obj.Key,
		})
		s.Require().NoError(err)
	}

	_, err = s.cli.DeleteBucket(&s3.DeleteBucketInput{
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

func fetchURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func uploadToURL(ctx context.Context, url string, payload []byte) error {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "multipart/form-data")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.Errorf("status code 200 != %d", resp.StatusCode)
	}
	return nil
}
