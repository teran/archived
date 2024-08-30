package service

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"

	"github.com/teran/archived/models"
	blobRepoMock "github.com/teran/archived/repositories/blob/mock"
	"github.com/teran/archived/repositories/metadata"
	mdRepoMock "github.com/teran/archived/repositories/metadata/mock"
)

const defaultNamespace = "default"

func (s *serviceTestSuite) TestCreateNamespace() {
	// Happy path
	s.mdRepoMock.On("CreateNamespace", "test-namespace").Return(nil).Once()

	err := s.svc.CreateNamespace(s.ctx, "test-namespace")
	s.Require().NoError(err)

	// return error
	s.mdRepoMock.On("CreateNamespace", "test-namespace").Return(errors.New("test error")).Once()

	err = s.svc.CreateNamespace(s.ctx, "test-namespace")
	s.Require().Error(err)
	s.Require().Equal("test error", err.Error())
}

func (s *serviceTestSuite) TestListNamespaces() {
	// Happy path
	s.mdRepoMock.On("ListNamespaces").Return([]string{
		"namespace1", "namespace2",
	}, nil).Once()

	containers, err := s.svc.ListNamespaces(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"namespace1", "namespace2",
	}, containers)

	// return error
	s.mdRepoMock.On("ListNamespaces").Return([]string(nil), errors.New("test error")).Once()

	_, err = s.svc.ListNamespaces(s.ctx)
	s.Require().Error(err)
	s.Require().Equal("test error", err.Error())
}

func (s *serviceTestSuite) TestRenameNamespace() {
	// Happy path
	s.mdRepoMock.On("RenameNamespace", "old-name", "new-name").Return(nil).Once()

	err := s.svc.RenameNamespace(s.ctx, "old-name", "new-name")
	s.Require().NoError(err)

	// return error
	s.mdRepoMock.On("RenameNamespace", "old-name", "new-name").Return(errors.New("test error")).Once()

	err = s.svc.RenameNamespace(s.ctx, "old-name", "new-name")
	s.Require().Error(err)
	s.Require().Equal("test error", err.Error())
}

func (s *serviceTestSuite) TestDeleteNamespace() {
	s.mdRepoMock.On("DeleteNamespace", "namespace1").Return(nil).Once()

	err := s.svc.DeleteNamespace(s.ctx, "namespace1")
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TestCreateContainer() {
	// Happy path
	s.mdRepoMock.On("CreateContainer", defaultNamespace, "container").Return(nil).Once()

	err := s.svc.CreateContainer(s.ctx, defaultNamespace, "container")
	s.Require().NoError(err)

	// return error
	s.mdRepoMock.On("CreateContainer", defaultNamespace, "container").Return(errors.New("test error")).Once()

	err = s.svc.CreateContainer(s.ctx, defaultNamespace, "container")
	s.Require().Error(err)
	s.Require().Equal("test error", err.Error())
}

func (s *serviceTestSuite) TestMoveContainer() {
	// Happy path
	s.mdRepoMock.On("RenameContainer", defaultNamespace, "container", "new-namespace", "container").Return(nil).Once()

	err := s.svc.MoveContainer(s.ctx, defaultNamespace, "container", "new-namespace")
	s.Require().NoError(err)

	// return error
	s.mdRepoMock.On("RenameContainer", defaultNamespace, "container", "new-namespace", "container").Return(errors.New("test error")).Once()

	err = s.svc.MoveContainer(s.ctx, defaultNamespace, "container", "new-namespace")
	s.Require().Error(err)
	s.Require().Equal("test error", err.Error())
}

func (s *serviceTestSuite) TestRenameContainer() {
	// Happy path
	s.mdRepoMock.On("RenameContainer", defaultNamespace, "old-name", defaultNamespace, "new-name").Return(nil).Once()

	err := s.svc.RenameContainer(s.ctx, defaultNamespace, "old-name", "new-name")
	s.Require().NoError(err)

	// return error
	s.mdRepoMock.On("RenameContainer", defaultNamespace, "old-name", defaultNamespace, "new-name").Return(errors.New("test error")).Once()

	err = s.svc.RenameContainer(s.ctx, defaultNamespace, "old-name", "new-name")
	s.Require().Error(err)
	s.Require().Equal("test error", err.Error())
}

func (s *serviceTestSuite) TestDeleteContainer() {
	s.mdRepoMock.On("DeleteContainer", defaultNamespace, "container").Return(nil).Once()

	err := s.svc.DeleteContainer(s.ctx, defaultNamespace, "container")
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TestCreateVersion() {
	s.mdRepoMock.On("CreateVersion", defaultNamespace, "container").Return("versionID", nil).Once()

	id, err := s.svc.CreateVersion(s.ctx, defaultNamespace, "container")
	s.Require().NoError(err)
	s.Require().Equal("versionID", id)
}

func (s *serviceTestSuite) TestPublishVersion() {
	s.mdRepoMock.On("MarkVersionPublished", defaultNamespace, "container", "version").Return(nil).Once()

	err := s.svc.PublishVersion(s.ctx, defaultNamespace, "container", "version")
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TestDeleteVersion() {
	s.mdRepoMock.On("DeleteVersion", defaultNamespace, "test_container", "test_version").Return(nil).Once()

	err := s.svc.DeleteVersion(s.ctx, defaultNamespace, "test_container", "test_version")
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TestAddObject() {
	s.mdRepoMock.On("CreateObject", defaultNamespace, "container", "versionID", "key", "cas_key").Return(nil).Once()

	err := s.svc.AddObject(s.ctx, defaultNamespace, "container", "versionID", "key", "cas_key")
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TestAddObjectWithLeadingSlash() {
	s.mdRepoMock.On("CreateObject", defaultNamespace, "container", "versionID", "key", "cas_key").Return(nil).Once()

	err := s.svc.AddObject(s.ctx, defaultNamespace, "container", "versionID", "/key", "cas_key")
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TestDeleteObject() {
	s.mdRepoMock.On("DeleteObject", defaultNamespace, "test-container", "test-version", []string{"test-key"}).Return(nil).Once()

	err := s.svc.DeleteObject(s.ctx, defaultNamespace, "test-container", "test-version", "test-key")
	s.Require().NoError(err)
}

func (s *serviceTestSuite) TestListContainers() {
	// Happy path
	s.mdRepoMock.On("ListContainers", defaultNamespace).Return([]models.Container{
		{Name: "container1"},
		{Name: "container2"},
	}, nil).Once()

	containers, err := s.svc.ListContainers(s.ctx, defaultNamespace)
	s.Require().NoError(err)
	s.Require().Equal([]models.Container{
		{Name: "container1"},
		{Name: "container2"},
	}, containers)

	// return error
	s.mdRepoMock.On("ListContainers", defaultNamespace).Return([]models.Container(nil), errors.New("test error")).Once()

	_, err = s.svc.ListContainers(s.ctx, defaultNamespace)
	s.Require().Error(err)
	s.Require().Equal("test error", err.Error())
}


func (s *serviceTestSuite) TestListContainersByPage() {
	s.mdRepoMock.On("ListContainersByPage", defaultNamespace, uint64(300), uint64(50)).Return(uint64(200), []models.Container{
		{Name: "container1"},
		{Name: "container2"},
	}, nil).Once()

	total, containers, err := s.svc.ListContainersByPage(s.ctx, defaultNamespace, 7)
	s.Require().NoError(err)
	s.Require().Equal(uint64(4), total)
	s.Require().Equal([]models.Container{
		{Name: "container1"},
		{Name: "container2"},
	}, containers)
}

func (s *serviceTestSuite) TestListPublishedVersions() {
	s.mdRepoMock.On("ListPublishedVersionsByContainer", defaultNamespace, "container").Return([]models.Version{
		{Name: "version1"},
		{Name: "version2"},
	}, nil).Once()

	versions, err := s.svc.ListPublishedVersions(s.ctx, defaultNamespace, "container")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{
		{Name: "version1"},
		{Name: "version2"},
	}, versions)
}

func (s *serviceTestSuite) TestListPublishedVersionsByPage() {
	s.mdRepoMock.
		On("ListPublishedVersionsByContainerAndPage", defaultNamespace, "container", uint64(450), uint64(50)).
		Return(uint64(1000), []models.Version{
			{Name: "version1"},
			{Name: "version2"},
		}, nil).Once()

	total, versions, err := s.svc.ListPublishedVersionsByPage(s.ctx, defaultNamespace, "container", 10)
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{
		{Name: "version1"},
		{Name: "version2"},
	}, versions)
	s.Require().Equal(uint64(20), total)
}

func (s *serviceTestSuite) TestListObjects() {
	// Happy path
	s.mdRepoMock.On("ListObjects", defaultNamespace, "container", "versionID", uint64(0), uint64(1000)).Return(uint64(100), []string{
		"object1", "object2",
	}, nil).Once()

	objects, err := s.svc.ListObjects(s.ctx, defaultNamespace, "container", "versionID")
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"object1", "object2",
	}, objects)

	// return error
	s.mdRepoMock.On("ListObjects", defaultNamespace, "container", "versionID", uint64(0), uint64(1000)).Return(uint64(100), []string(nil), errors.New("test error")).Once()

	_, err = s.svc.ListObjects(s.ctx, defaultNamespace, "container", "versionID")
	s.Require().Error(err)
	s.Require().Equal("test error", err.Error())
}

func (s *serviceTestSuite) TestGetObjectURL() {
	// Happy path
	s.mdRepoMock.On("GetBlobKeyByObject", defaultNamespace, "container", "versionID", "key").Return("deadbeef", nil).Once()
	s.blobRepoMock.On("GetBlobURL", "deadbeef").Return("url", nil).Once()

	url, err := s.svc.GetObjectURL(s.ctx, defaultNamespace, "container", "versionID", "key")
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

func (s *serviceTestSuite) TestListObjectsByLatestVersion() {
	s.mdRepoMock.On("GetLatestPublishedVersionByContainer", defaultNamespace, "container1").Return("versionID", nil).Once()
	s.mdRepoMock.On("ListObjects", defaultNamespace, "container1", "versionID", uint64(0), uint64(50)).Return(uint64(100), []string{"obj1", "obj2"}, nil).Once()

	_, objects, err := s.svc.ListObjectsByPage(s.ctx, defaultNamespace, "container1", "latest", 1)
	s.Require().NoError(err)
	s.Require().Equal([]string{"obj1", "obj2"}, objects)
}

func (s *serviceTestSuite) TestGetObjectURLWithLatestVersion() {
	s.mdRepoMock.On("GetLatestPublishedVersionByContainer", defaultNamespace, "container12").Return("versionID", nil).Once()
	s.mdRepoMock.On("GetBlobKeyByObject", defaultNamespace, "container12", "versionID", "key").Return("deadbeef", nil).Once()
	s.blobRepoMock.On("GetBlobURL", "deadbeef").Return("url", nil).Once()

	url, err := s.svc.GetObjectURL(s.ctx, defaultNamespace, "container12", "latest", "key")
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

	s.svc = newSvc(s.mdRepoMock, s.blobRepoMock, 50, 50, 50)
}

func (s *serviceTestSuite) TearDownTest() {
	s.mdRepoMock.AssertExpectations(s.T())
	s.blobRepoMock.AssertExpectations(s.T())
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, &serviceTestSuite{})
}
