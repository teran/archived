package postgresql

import (
	"time"

	"github.com/teran/archived/models"
	"github.com/teran/archived/repositories/metadata"
)

func (s *postgreSQLRepositoryTestSuite) TestVersionsOperations() {
	// CreateContainer (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Twice()

	// CreateVersion (created_at) - for each call
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Once()
	s.tp.On("Now").Return("2024-07-07T11:12:13Z").Once()

	err := s.repo.CreateContainer(s.ctx, defaultNamespace, "container1", -1)
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "container2", -1)
	s.Require().NoError(err)

	vName, err := s.repo.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal("20240707101112", vName)

	vName, err = s.repo.CreateVersion(s.ctx, defaultNamespace, "container2")
	s.Require().NoError(err)
	s.Require().Equal("20240707111213", vName)

	listC1, err := s.repo.ListAllVersionsByContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{
		{
			Name:      "20240707101112",
			CreatedAt: time.Date(2024, 7, 7, 10, 11, 12, 0, time.UTC),
		},
	}, listC1)

	listC2, err := s.repo.ListAllVersionsByContainer(s.ctx, defaultNamespace, "container2")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{
		{
			Name:      "20240707111213",
			CreatedAt: time.Date(2024, 7, 7, 11, 12, 13, 0, time.UTC),
		},
	}, listC2)
}

func (s *postgreSQLRepositoryTestSuite) TestPublishVersion() {
	// CreateContainer (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:10Z").Once()

	// CreateVersion (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Once()

	err := s.repo.CreateContainer(s.ctx, defaultNamespace, "container1", -1)
	s.Require().NoError(err)

	version, err := s.repo.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)

	list, err := s.repo.ListAllVersionsByContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{
		{
			Name:      "20240707101112",
			CreatedAt: time.Date(2024, 7, 7, 10, 11, 12, 0, time.UTC),
		},
	}, list)

	list, err = s.repo.ListPublishedVersionsByContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{}, list)

	err = s.repo.MarkVersionPublished(s.ctx, defaultNamespace, "container1", version)
	s.Require().NoError(err)

	list, err = s.repo.ListAllVersionsByContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{
		{
			Name:        "20240707101112",
			IsPublished: true,
			CreatedAt:   time.Date(2024, 7, 7, 10, 11, 12, 0, time.UTC),
		},
	}, list)

	list, err = s.repo.ListPublishedVersionsByContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{
		{
			Name:        "20240707101112",
			IsPublished: true,
			CreatedAt:   time.Date(2024, 7, 7, 10, 11, 12, 0, time.UTC),
		},
	}, list)
}

func (s *postgreSQLRepositoryTestSuite) TestPublishVersionErrors() {
	// Not existent container
	err := s.repo.MarkVersionPublished(s.ctx, defaultNamespace, "not-existent", "version")
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)

	// Not existent version
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Once()

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "container1", -1)
	s.Require().NoError(err)

	err = s.repo.MarkVersionPublished(s.ctx, defaultNamespace, "not-existent", "version")
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)
}

func (s *postgreSQLRepositoryTestSuite) TestListAllVersionsByContainerErrors() {
	_, err := s.repo.ListAllVersionsByContainer(s.ctx, defaultNamespace, "test-container")
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)
}

func (s *postgreSQLRepositoryTestSuite) TestListObjectsErrorsNotExistentContainer() {
	_, _, err := s.repo.ListObjects(s.ctx, defaultNamespace, "test-container", "2024-01-02T03:04:05Z", 0, 100)
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)
}

func (s *postgreSQLRepositoryTestSuite) TestListObjectsErrorsNotExistentVersion() {
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Once()

	err := s.repo.CreateContainer(s.ctx, defaultNamespace, "test-container", -1)
	s.Require().NoError(err)

	_, _, err = s.repo.ListObjects(s.ctx, defaultNamespace, "test-container", "2024-01-02T03:04:05Z", 0, 100)
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)
}

func (s *postgreSQLRepositoryTestSuite) TestVersionsPagination() {
	// CreateContainer (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Once()

	// CreateVersion (created_at) - for each call
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Once()
	s.tp.On("Now").Return("2024-07-07T10:11:13Z").Once()
	s.tp.On("Now").Return("2024-07-07T10:11:14Z").Once()

	err := s.repo.CreateContainer(s.ctx, defaultNamespace, "container1", -1)
	s.Require().NoError(err)

	version1, err := s.repo.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)

	err = s.repo.MarkVersionPublished(s.ctx, defaultNamespace, "container1", version1)
	s.Require().NoError(err)

	_, err = s.repo.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)

	version3, err := s.repo.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)

	err = s.repo.MarkVersionPublished(s.ctx, defaultNamespace, "container1", version3)
	s.Require().NoError(err)

	total, listByPage, err := s.repo.ListPublishedVersionsByContainerAndPage(s.ctx, defaultNamespace, "container1", 0, 5)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), total)
	s.Require().Equal([]models.Version{
		{
			Name:        version3,
			IsPublished: true,
			CreatedAt:   time.Date(2024, 7, 7, 10, 11, 14, 0, time.UTC),
		},
		{
			Name:        version1,
			IsPublished: true,
			CreatedAt:   time.Date(2024, 7, 7, 10, 11, 12, 0, time.UTC),
		},
	}, listByPage)

	total, listByPage, err = s.repo.ListPublishedVersionsByContainerAndPage(s.ctx, defaultNamespace, "container1", 1, 2)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), total)
	s.Require().Equal([]models.Version{
		{
			Name:        version1,
			IsPublished: true,
			CreatedAt:   time.Date(2024, 7, 7, 10, 11, 12, 0, time.UTC),
		},
	}, listByPage)
}

func (s *postgreSQLRepositoryTestSuite) TestDeleteVersion() {
	// CreateContainer (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Twice()

	// CreateVersion (created_at) - for each call
	s.tp.On("Now").Return("2024-07-07T10:11:13Z").Once()
	s.tp.On("Now").Return("2024-07-07T10:11:14Z").Once()
	s.tp.On("Now").Return("2024-07-07T10:11:15Z").Once()
	s.tp.On("Now").Return("2024-07-07T10:11:16Z").Once()

	// CreateBLOB (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:16Z").Once()

	// CreateObject (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:16Z").Once()

	err := s.repo.CreateContainer(s.ctx, defaultNamespace, "container1", -1)
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "container2", -1)
	s.Require().NoError(err)

	version1, err := s.repo.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)

	version2, err := s.repo.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)

	version3, err := s.repo.CreateVersion(s.ctx, defaultNamespace, "container2")
	s.Require().NoError(err)

	version4, err := s.repo.CreateVersion(s.ctx, defaultNamespace, "container2")
	s.Require().NoError(err)

	err = s.repo.CreateBLOB(s.ctx, "deadbeef", 10, "application/octet-stream")
	s.Require().NoError(err)

	err = s.repo.CreateObject(s.ctx, defaultNamespace, "container2", version4, "testkey", "deadbeef")
	s.Require().NoError(err)

	versions1, err := s.repo.ListAllVersionsByContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{
		{
			Name:      version2,
			CreatedAt: time.Date(2024, 7, 7, 10, 11, 14, 0, time.UTC),
		},
		{
			Name:      version1,
			CreatedAt: time.Date(2024, 7, 7, 10, 11, 13, 0, time.UTC),
		},
	}, versions1)

	versions2, err := s.repo.ListAllVersionsByContainer(s.ctx, defaultNamespace, "container2")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{
		{
			Name:      version4,
			CreatedAt: time.Date(2024, 7, 7, 10, 11, 16, 0, time.UTC),
		},
		{
			Name:      version3,
			CreatedAt: time.Date(2024, 7, 7, 10, 11, 15, 0, time.UTC),
		},
	}, versions2)

	err = s.repo.DeleteVersion(s.ctx, defaultNamespace, "container1", version1)
	s.Require().NoError(err)

	versions1, err = s.repo.ListAllVersionsByContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{
		{
			Name:      version2,
			CreatedAt: time.Date(2024, 7, 7, 10, 11, 14, 0, time.UTC),
		},
	}, versions1)

	versions2, err = s.repo.ListAllVersionsByContainer(s.ctx, defaultNamespace, "container2")
	s.Require().NoError(err)
	s.Require().Equal([]models.Version{
		{
			Name:      version4,
			CreatedAt: time.Date(2024, 7, 7, 10, 11, 16, 0, time.UTC),
		},
		{
			Name:      version3,
			CreatedAt: time.Date(2024, 7, 7, 10, 11, 15, 0, time.UTC),
		},
	}, versions2)
}

func (s *postgreSQLRepositoryTestSuite) TestGetLatestPublishedVersionByContainer() {
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Times(3)
	s.tp.On("Now").Return("2024-07-07T10:11:13Z").Once()
	s.tp.On("Now").Return("2024-07-07T10:11:14Z").Once()
	s.tp.On("Now").Return("2024-07-07T10:11:15Z").Once()
	s.tp.On("Now").Return("2024-07-07T10:11:16Z").Once()

	err := s.repo.CreateContainer(s.ctx, defaultNamespace, "container1", -1)
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "container2", -1)
	s.Require().NoError(err)

	_, err = s.repo.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)

	version2, err := s.repo.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)

	err = s.repo.MarkVersionPublished(s.ctx, defaultNamespace, "container1", version2)
	s.Require().NoError(err)

	_, err = s.repo.CreateVersion(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)

	_, err = s.repo.CreateVersion(s.ctx, defaultNamespace, "container2")
	s.Require().NoError(err)

	_, err = s.repo.CreateVersion(s.ctx, defaultNamespace, "container2")
	s.Require().NoError(err)

	versionName, err := s.repo.GetLatestPublishedVersionByContainer(s.ctx, defaultNamespace, "container1")
	s.Require().NoError(err)
	s.Require().Equal(version2, versionName)

	_, err = s.repo.GetLatestPublishedVersionByContainer(s.ctx, defaultNamespace, "container2")
	s.Require().Error(err)
	s.Require().Equal("not found", err.Error())
}
