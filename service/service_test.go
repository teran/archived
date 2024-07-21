package service

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"

	blobRepoMock "github.com/teran/archived/repositories/blob/mock"
	"github.com/teran/archived/repositories/metadata"
	mdRepoMock "github.com/teran/archived/repositories/metadata/mock"
)

func (s *serviceTestSuite) TestCreateContainer() {
	// Happy path
	s.mdRepoMock.On("CreateContainer", "container").Return(nil).Once()

	err := s.svc.CreateContainer(s.ctx, "container")
	s.Require().NoError(err)

	// return error
	s.mdRepoMock.On("CreateContainer", "container").Return(errors.New("test error")).Once()

	err = s.svc.CreateContainer(s.ctx, "container")
	s.Require().Error(err)
	s.Require().Equal("error creating container: test error", err.Error())
}

func (s *serviceTestSuite) TestDeleteContainer() {
	s.mdRepoMock.On("DeleteContainer", "container").Return(nil).Once()

	err := s.svc.DeleteContainer(s.ctx, "container")
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TestCreateVersion() {
	s.mdRepoMock.On("CreateVersion", "container").Return("versionID", nil).Once()

	id, err := s.svc.CreateVersion(s.ctx, "container")
	s.Require().NoError(err)
	s.Require().Equal("versionID", id)
}

func (s *serviceTestSuite) TestPublishVersion() {
	s.mdRepoMock.On("MarkVersionPublished", "container", "version").Return(nil).Once()

	err := s.svc.PublishVersion(s.ctx, "container", "version")
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TestDeleteVersion() {
	s.Require().PanicsWithValue("not implemented", func() {
		s.svc.DeleteVersion(s.ctx, "container", "version")
	})
}

func (s *serviceTestSuite) TestAddObject() {
	s.mdRepoMock.On("CreateObject", "container", "versionID", "key", "cas_key").Return(nil).Once()

	err := s.svc.AddObject(s.ctx, "container", "versionID", "key", "cas_key")
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TestDeleteObject() {
	s.Require().PanicsWithValue("not implemented", func() {
		s.svc.DeleteObject(s.ctx, "container", "version", "key")
	})
}

func (s *serviceTestSuite) TestListContainers() {
	// Happy path
	s.mdRepoMock.On("ListContainers").Return([]string{
		"container1", "container2",
	}, nil).Once()

	containers, err := s.svc.ListContainers(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"container1", "container2",
	}, containers)

	// return error
	s.mdRepoMock.On("ListContainers").Return([]string(nil), errors.New("test error")).Once()

	_, err = s.svc.ListContainers(s.ctx)
	s.Require().Error(err)
	s.Require().Equal("test error", err.Error())
}

func (s *serviceTestSuite) TestListPublishedVersions() {
	s.mdRepoMock.On("ListPublishedVersionsByContainer", "container").Return([]string{
		"version1", "version2",
	}, nil).Once()

	versions, err := s.svc.ListPublishedVersions(s.ctx, "container")
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"version1", "version2",
	}, versions)
}

func (s *serviceTestSuite) TestListPublishedVersionsByPage() {
	s.mdRepoMock.
		On("ListPublishedVersionsByContainerAndPage", "container", uint64(450), uint64(50)).
		Return(uint64(1000), []string{
			"version1", "version2",
		}, nil).Once()

	total, versions, err := s.svc.ListPublishedVersionsByPage(s.ctx, "container", 10)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"version1", "version2",
	}, versions)
	s.Require().Equal(uint64(20), total)
}

func (s *serviceTestSuite) TestListObjects() {
	// Happy path
	s.mdRepoMock.On("ListObjects", "container", "versionID", uint64(0), uint64(1000)).Return([]string{
		"object1", "object2",
	}, nil).Once()

	objects, err := s.svc.ListObjects(s.ctx, "container", "versionID")
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"object1", "object2",
	}, objects)

	// return error
	s.mdRepoMock.On("ListObjects", "container", "versionID", uint64(0), uint64(1000)).Return([]string(nil), errors.New("test error")).Once()

	_, err = s.svc.ListObjects(s.ctx, "container", "versionID")
	s.Require().Error(err)
	s.Require().Equal("test error", err.Error())
}

func (s *serviceTestSuite) TestGetObjectURL() {
	// Happy path
	s.mdRepoMock.On("GetBlobKeyByObject", "container", "versionID", "key").Return("deadbeef", nil).Once()
	s.blobRepoMock.On("GetBlobURL", "deadbeef").Return("url", nil).Once()

	url, err := s.svc.GetObjectURL(s.ctx, "container", "versionID", "key")
	s.Require().NoError(err)
	s.Require().Equal("url", url)
}

func (s *serviceTestSuite) TestEnsureBLOBPresenceOrGetUploadURL() {
	// Blob exists
	s.mdRepoMock.On("EnsureBlobKey", "checksum", uint64(1234)).Return(nil).Once()

	url, err := s.svc.EnsureBLOBPresenceOrGetUploadURL(s.ctx, "checksum", 1234)
	s.Require().NoError(err)
	s.Require().Equal("", url)

	// Blob doesn't exist
	s.mdRepoMock.On("EnsureBlobKey", "checksum", uint64(1234)).Return(metadata.ErrNotFound).Once()
	s.blobRepoMock.On("PutBlobURL", "checksum").Return("https://example.com", nil).Once()
	s.mdRepoMock.On("CreateBLOB", "checksum", uint64(1234), "application/octet-stream").Return(nil).Once()

	url, err = s.svc.EnsureBLOBPresenceOrGetUploadURL(s.ctx, "checksum", 1234)
	s.Require().NoError(err)
	s.Require().Equal("https://example.com", url)
}

// Definitions
type serviceTestSuite struct {
	suite.Suite

	ctx          context.Context
	svc          *service
	mdRepoMock   *mdRepoMock.Mock
	blobRepoMock *blobRepoMock.Mock
}

func (s *serviceTestSuite) SetupTest() {
	s.ctx = context.Background()

	s.mdRepoMock = mdRepoMock.New()
	s.blobRepoMock = blobRepoMock.New()

	s.svc = newSvc(s.mdRepoMock, s.blobRepoMock)
}

func (s *serviceTestSuite) TearDownTest() {
	s.mdRepoMock.AssertExpectations(s.T())
	s.blobRepoMock.AssertExpectations(s.T())
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, &serviceTestSuite{})
}
