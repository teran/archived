package memcache

import (
	"context"
	"testing"
	"time"

	memcacheCli "github.com/bradfitz/gomemcache/memcache"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	emodels "github.com/teran/archived/exporter/models"
	"github.com/teran/archived/models"
	"github.com/teran/archived/repositories/metadata"
	repoM "github.com/teran/archived/repositories/metadata/mock"
	memcacheApp "github.com/teran/go-docker-testsuite/applications/memcache"
)

const defaultNamespace = "default"

func init() {
	log.SetLevel(log.TraceLevel)
}

// Cached methods ...
func (s *memcacheTestSuite) TestListNamespaces() {
	s.repoMock.On("ListNamespaces").Return([]string{"namespace1", "namespace2"}, nil).Once()

	containers, err := s.cache.ListNamespaces(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{"namespace1", "namespace2"}, containers)

	containers, err = s.cache.ListNamespaces(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{"namespace1", "namespace2"}, containers)
}

func (s *memcacheTestSuite) TestListContainers() {
	s.repoMock.On("ListContainers", defaultNamespace).Return([]models.Container{{Name: "container1"}}, nil).Once()

	containers, err := s.cache.ListContainers(s.ctx, defaultNamespace)
	s.Require().NoError(err)
	s.Require().Equal([]models.Container{{Name: "container1"}}, containers)

	containers, err = s.cache.ListContainers(s.ctx, defaultNamespace)
	s.Require().NoError(err)
	s.Require().Equal([]models.Container{{Name: "container1"}}, containers)
}

func (s *memcacheTestSuite) TestListContainersError() {
	s.repoMock.On("ListContainers", defaultNamespace).Return([]models.Container{}, errors.New("some error")).Once()

	_, err := s.cache.ListContainers(s.ctx, defaultNamespace)
	s.Require().Error(err)
	s.Require().Equal("some error", err.Error())
}

func (s *memcacheTestSuite) TestListContainersByPage() {
	s.repoMock.On("ListContainersByPage", defaultNamespace, uint64(0), uint64(15)).Return(uint64(500), []models.Container{{Name: "container1"}}, nil).Once()

	total, containers, err := s.cache.ListContainersByPage(s.ctx, defaultNamespace, 0, 15)
	s.Require().NoError(err)
	s.Require().Equal([]models.Container{{Name: "container1"}}, containers)
	s.Require().Equal(uint64(500), total)

	total, containers, err = s.cache.ListContainersByPage(s.ctx, defaultNamespace, 0, 15)
	s.Require().NoError(err)
	s.Require().Equal([]models.Container{{Name: "container1"}}, containers)
	s.Require().Equal(uint64(500), total)
}

func (s *memcacheTestSuite) TestGetLatestPublishedVersionByContainer() {
	s.repoMock.On("GetLatestPublishedVersionByContainer", defaultNamespace, "test-container").Return("test-version", nil).Once()

	version, err := s.cache.GetLatestPublishedVersionByContainer(s.ctx, defaultNamespace, "test-container")
	s.Require().NoError(err)
	s.Require().Equal("test-version", version)

	version, err = s.cache.GetLatestPublishedVersionByContainer(s.ctx, defaultNamespace, "test-container")
	s.Require().NoError(err)
	s.Require().Equal("test-version", version)
}

func (s *memcacheTestSuite) TestGetLatestPublishedVersionByContainerError() {
	s.repoMock.On("GetLatestPublishedVersionByContainer", defaultNamespace, "test-container").Return("", errors.New("some error")).Once()

	_, err := s.cache.GetLatestPublishedVersionByContainer(s.ctx, defaultNamespace, "test-container")
	s.Require().Error(err)
	s.Require().Equal("some error", err.Error())
}

func (s *memcacheTestSuite) TestListAllVersionsByContainer() {
	s.repoMock.On("ListAllVersionsByContainer", defaultNamespace, "test-container").Return([]models.Version{{Name: "test-version"}}, nil).Once()

	versions, err := s.cache.ListAllVersionsByContainer(s.ctx, defaultNamespace, "test-container")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{{Name: "test-version"}}, versions)

	versions, err = s.cache.ListAllVersionsByContainer(s.ctx, defaultNamespace, "test-container")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{{Name: "test-version"}}, versions)
}

func (s *memcacheTestSuite) TestListAllVersionsByContainerError() {
	s.repoMock.On("ListAllVersionsByContainer", defaultNamespace, "test-container").Return([]models.Version{}, errors.New("some error")).Once()

	_, err := s.cache.ListAllVersionsByContainer(s.ctx, defaultNamespace, "test-container")
	s.Require().Error(err)
	s.Require().Equal("some error", err.Error())
}

func (s *memcacheTestSuite) TestListPublishedVersionsByContainer() {
	s.repoMock.On("ListPublishedVersionsByContainer", defaultNamespace, "test-container").Return([]models.Version{{Name: "test-version"}}, nil).Once()

	versions, err := s.cache.ListPublishedVersionsByContainer(s.ctx, defaultNamespace, "test-container")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{{Name: "test-version"}}, versions)

	versions, err = s.cache.ListPublishedVersionsByContainer(s.ctx, defaultNamespace, "test-container")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{{Name: "test-version"}}, versions)
}

func (s *memcacheTestSuite) TestListPublishedVersionsByContainerError() {
	s.repoMock.On("ListPublishedVersionsByContainer", defaultNamespace, "test-container").Return([]models.Version{}, errors.New("some error")).Once()

	_, err := s.cache.ListPublishedVersionsByContainer(s.ctx, defaultNamespace, "test-container")
	s.Require().Error(err)
	s.Require().Equal("some error", err.Error())
}

func (s *memcacheTestSuite) TestListPublishedVersionsByContainerAndPage() {
	s.repoMock.On("ListPublishedVersionsByContainerAndPage", defaultNamespace, "test-container", uint64(0), uint64(15)).Return(uint64(500), []models.Version{{Name: "test-version"}}, nil).Once()

	total, versions, err := s.cache.ListPublishedVersionsByContainerAndPage(s.ctx, defaultNamespace, "test-container", 0, 15)
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{{Name: "test-version"}}, versions)
	s.Require().Equal(uint64(500), total)

	total, versions, err = s.cache.ListPublishedVersionsByContainerAndPage(s.ctx, defaultNamespace, "test-container", 0, 15)
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{{Name: "test-version"}}, versions)
	s.Require().Equal(uint64(500), total)
}

func (s *memcacheTestSuite) TestListPublishedVersionsByContainerAndPageError() {
	s.repoMock.On("ListPublishedVersionsByContainerAndPage", defaultNamespace, "test-container", uint64(0), uint64(15)).Return(uint64(0), []models.Version{}, errors.New("some error")).Once()

	_, _, err := s.cache.ListPublishedVersionsByContainerAndPage(s.ctx, defaultNamespace, "test-container", 0, 15)
	s.Require().Error(err)
	s.Require().Equal("some error", err.Error())
}

func (s *memcacheTestSuite) TestListObjects() {
	s.repoMock.On("ListObjects", defaultNamespace, "test-container", "test-version", uint64(0), uint64(30)).Return(uint64(500), []string{"obj1"}, nil).Once()

	total, objects, err := s.cache.ListObjects(s.ctx, defaultNamespace, "test-container", "test-version", 0, 30)
	s.Require().NoError(err)
	s.Require().Equal([]string{"obj1"}, objects)
	s.Require().Equal(uint64(500), total)

	total, objects, err = s.cache.ListObjects(s.ctx, defaultNamespace, "test-container", "test-version", 0, 30)
	s.Require().NoError(err)
	s.Require().Equal([]string{"obj1"}, objects)
	s.Require().Equal(uint64(500), total)
}

func (s *memcacheTestSuite) TestListObjectsError() {
	s.repoMock.On("ListObjects", defaultNamespace, "test-container", "test-version", uint64(0), uint64(30)).Return(uint64(0), []string{}, errors.New("some error")).Once()

	_, _, err := s.cache.ListObjects(s.ctx, defaultNamespace, "test-container", "test-version", 0, 30)
	s.Require().Error(err)
	s.Require().Equal("some error", err.Error())
}

func (s *memcacheTestSuite) TestGetBlobKeyByObject() {
	s.repoMock.On("GetBlobKeyByObject", defaultNamespace, "container", "version", "key").Return("deadbeef", nil).Once()

	casKey, err := s.cache.GetBlobKeyByObject(s.ctx, defaultNamespace, "container", "version", "key")
	s.Require().NoError(err)
	s.Require().Equal("deadbeef", casKey)

	casKey, err = s.cache.GetBlobKeyByObject(s.ctx, defaultNamespace, "container", "version", "key")
	s.Require().NoError(err)
	s.Require().Equal("deadbeef", casKey)
}

func (s *memcacheTestSuite) TestGetBlobKeyByObjectError() {
	s.repoMock.On("GetBlobKeyByObject", defaultNamespace, "container", "version", "key").Return("", errors.New("some error")).Once()

	_, err := s.cache.GetBlobKeyByObject(s.ctx, defaultNamespace, "container", "version", "key")
	s.Require().Error(err)
	s.Require().Equal("some error", err.Error())
}

// Non-cached methods ...
func (s *memcacheTestSuite) TestCreateNamespace() {
	s.repoMock.On("CreateNamespace", "namespace1").Return(nil).Twice()

	err := s.cache.CreateNamespace(s.ctx, "namespace1")
	s.Require().NoError(err)

	err = s.cache.CreateNamespace(s.ctx, "namespace1")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestRenameNamespace() {
	s.repoMock.On("RenameNamespace", "old-name", "new-name").Return(nil).Twice()

	err := s.cache.RenameNamespace(s.ctx, "old-name", "new-name")
	s.Require().NoError(err)

	err = s.cache.RenameNamespace(s.ctx, "old-name", "new-name")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestDeleteNamespace() {
	s.repoMock.On("DeleteNamespace", "namespace1").Return(nil).Twice()

	err := s.cache.DeleteNamespace(s.ctx, "namespace1")
	s.Require().NoError(err)

	err = s.cache.DeleteNamespace(s.ctx, "namespace1")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestCreateContainer() {
	s.repoMock.On("CreateContainer", defaultNamespace, "container1").Return(nil).Twice()

	err := s.cache.CreateContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)

	err = s.cache.CreateContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestRenameContainer() {
	s.repoMock.On("RenameContainer", defaultNamespace, "old-name", "new-namespace", "new-name").Return(nil).Twice()

	err := s.cache.RenameContainer(s.ctx, defaultNamespace, "old-name", "new-namespace", "new-name")
	s.Require().NoError(err)

	err = s.cache.RenameContainer(s.ctx, defaultNamespace, "old-name", "new-namespace", "new-name")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestDeleteContainer() {
	s.repoMock.On("DeleteContainer", defaultNamespace, "container1").Return(nil).Twice()

	err := s.cache.DeleteContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)

	err = s.cache.DeleteContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestCreateVersion() {
	s.repoMock.On("CreateVersion", defaultNamespace, "container1").Return("test-version", nil).Twice()

	version, err := s.cache.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal("test-version", version)

	version, err = s.cache.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal("test-version", version)
}

func (s *memcacheTestSuite) TestListUnpublishedVersionsByContainer() {
	s.repoMock.On("ListUnpublishedVersionsByContainer", defaultNamespace, "container1").Return([]models.Version{{Name: "test-version"}}, nil).Twice()

	version, err := s.cache.ListUnpublishedVersionsByContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{{Name: "test-version"}}, version)

	version, err = s.cache.ListUnpublishedVersionsByContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{{Name: "test-version"}}, version)
}

func (s *memcacheTestSuite) TestMarkVersionPublished() {
	s.repoMock.On("MarkVersionPublished", defaultNamespace, "test-container", "test-version").Return(nil).Twice()

	err := s.cache.MarkVersionPublished(s.ctx, defaultNamespace, "test-container", "test-version")
	s.Require().NoError(err)

	err = s.cache.MarkVersionPublished(s.ctx, defaultNamespace, "test-container", "test-version")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestDeleteVersion() {
	s.repoMock.On("DeleteVersion", defaultNamespace, "test-container", "test-version").Return(nil).Twice()

	err := s.cache.DeleteVersion(s.ctx, defaultNamespace, "test-container", "test-version")
	s.Require().NoError(err)

	err = s.cache.DeleteVersion(s.ctx, defaultNamespace, "test-container", "test-version")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestCreateObject() {
	s.repoMock.On("CreateObject", defaultNamespace, "test-container", "test-version", "test-key", "deadbeef").Return(nil).Twice()

	err := s.cache.CreateObject(s.ctx, defaultNamespace, "test-container", "test-version", "test-key", "deadbeef")
	s.Require().NoError(err)

	err = s.cache.CreateObject(s.ctx, defaultNamespace, "test-container", "test-version", "test-key", "deadbeef")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestDeleteObject() {
	s.repoMock.On("DeleteObject", defaultNamespace, "test-container", "test-version", []string{"test-key"}).Return(nil).Twice()

	err := s.cache.DeleteObject(s.ctx, defaultNamespace, "test-container", "test-version", "test-key")
	s.Require().NoError(err)

	err = s.cache.DeleteObject(s.ctx, defaultNamespace, "test-container", "test-version", "test-key")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestRemapObject() {
	s.repoMock.On("RemapObject", defaultNamespace, "test-container", "test-version", "test-key", "deadbeef").Return(nil).Twice()

	err := s.cache.RemapObject(s.ctx, defaultNamespace, "test-container", "test-version", "test-key", "deadbeef")
	s.Require().NoError(err)

	err = s.cache.RemapObject(s.ctx, defaultNamespace, "test-container", "test-version", "test-key", "deadbeef")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestCreateBLOB() {
	s.repoMock.On("CreateBLOB", "deadbeef", uint64(325), "application/octet-stream").Return(nil).Twice()

	err := s.cache.CreateBLOB(s.ctx, "deadbeef", 325, "application/octet-stream")
	s.Require().NoError(err)

	err = s.cache.CreateBLOB(s.ctx, "deadbeef", 325, "application/octet-stream")
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) TestEnsureBlobKey() {
	s.repoMock.On("EnsureBlobKey", "key", uint64(325)).Return(nil).Twice()

	err := s.cache.EnsureBlobKey(s.ctx, "key", 325)
	s.Require().NoError(err)

	err = s.cache.EnsureBlobKey(s.ctx, "key", 325)
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) CountStats() {
	s.repoMock.On("CountStats").Return(&emodels.Stats{
		ContainersCount: 1,
	}, nil).Twice()

	stats, err := s.cache.CountStats(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(&emodels.Stats{ContainersCount: 1}, stats)

	stats, err = s.cache.CountStats(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(&emodels.Stats{ContainersCount: 1}, stats)
}

// Definitions ...
type memcacheTestSuite struct {
	suite.Suite

	ctx      context.Context
	cache    metadata.Repository
	repoMock *repoM.Mock

	memcachedApp memcacheApp.Memcache
}

func (s *memcacheTestSuite) SetupSuite() {
	s.ctx = context.TODO()

	var err error
	s.memcachedApp, err = memcacheApp.New(s.ctx)
	s.Require().NoError(err)
}

func (s *memcacheTestSuite) SetupTest() {
	s.repoMock = repoM.New()

	url, err := s.memcachedApp.GetEndpointAddress()
	s.Require().NoError(err)

	cli := memcacheCli.New(url)
	s.cache = New(cli, s.repoMock, 3*time.Second, s.T().Name())
}

func (s *memcacheTestSuite) TearDownTest() {
	s.repoMock.AssertExpectations(s.T())
}

func (s *memcacheTestSuite) TearDownSuite() {
	s.memcachedApp.Close(s.ctx)
}

func TestMemcacheTestSuite(t *testing.T) {
	suite.Run(t, &memcacheTestSuite{})
}
