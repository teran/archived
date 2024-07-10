package service

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"

	blobRepoMock "github.com/teran/archived/repositories/blob/mock"
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
	s.Require().PanicsWithValue("not implemented", func() {
		s.svc.CreateVersion(s.ctx, "container")
	})
}

func (s *serviceTestSuite) TestPublishVersion() {
	s.Require().PanicsWithValue("not implemented", func() {
		s.svc.PublishVersion(s.ctx, "container", "version")
	})
}

func (s *serviceTestSuite) TestDeleteVersion() {
	s.Require().PanicsWithValue("not implemented", func() {
		s.svc.DeleteVersion(s.ctx, "container", "version")
	})
}

func (s *serviceTestSuite) TestAddObject() {
	s.Require().PanicsWithValue("not implemented", func() {
		s.svc.AddObject(s.ctx, "container", "versionID", "key", nil)
	})
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

func (s *serviceTestSuite) TestListVersions() {
	s.mdRepoMock.On("ListPublishedVersionsByContainer", "container").Return([]string{
		"version1", "version2",
	}, nil).Once()

	versions, err := s.svc.ListVersions(s.ctx, "container")
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"version1", "version2",
	}, versions)
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
