package postgresql

import (
	"time"

	"github.com/teran/archived/models"
	"github.com/teran/archived/repositories/metadata"
)

func (s *postgreSQLRepositoryTestSuite) TestContainerOperations() {
	s.tp.On("Now").Return("2024-01-02T01:02:03Z").Times(4)

	err := s.repo.CreateNamespace(s.ctx, "new-namespace")
	s.Require().NoError(err)

	list, err := s.repo.ListContainers(s.ctx, defaultNamespace)
	s.Require().NoError(err)
	s.Require().Equal([]models.Container{}, list)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "test-container9", 86400*time.Second)
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "test-container5", -1)
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "test-container9", 3*time.Hour)
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrConflict, err)

	err = s.repo.SetContainerParameters(s.ctx, defaultNamespace, "test-container5", 1*time.Hour)
	s.Require().NoError(err)

	err = s.repo.SetContainerParameters(s.ctx, defaultNamespace, "test-container9", 2*time.Hour)
	s.Require().NoError(err)

	list, err = s.repo.ListContainers(s.ctx, defaultNamespace)
	s.Require().NoError(err)
	s.Require().Equal([]models.Container{
		{Name: "test-container5", CreatedAt: time.Date(2024, 1, 2, 1, 2, 3, 0, time.UTC), VersionsTTL: 3600 * time.Second},
		{Name: "test-container9", CreatedAt: time.Date(2024, 1, 2, 1, 2, 3, 0, time.UTC), VersionsTTL: 7200 * time.Second},
	}, list)

	err = s.repo.DeleteContainer(s.ctx, defaultNamespace, "test-container9")
	s.Require().NoError(err)

	err = s.repo.DeleteContainer(s.ctx, defaultNamespace, "not-existent")
	s.Require().NoError(err)

	list, err = s.repo.ListContainers(s.ctx, defaultNamespace)
	s.Require().NoError(err)
	s.Require().Equal([]models.Container{
		{Name: "test-container5", CreatedAt: time.Date(2024, 1, 2, 1, 2, 3, 0, time.UTC), VersionsTTL: 3600 * time.Second},
	}, list)

	err = s.repo.RenameContainer(s.ctx, defaultNamespace, "test-container5", "new-namespace", "and-then-there-was-the-one")
	s.Require().NoError(err)

	err = s.repo.RenameContainer(s.ctx, defaultNamespace, "not-existent", "new-namespace", "some-name")
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)

	list, err = s.repo.ListContainers(s.ctx, defaultNamespace)
	s.Require().NoError(err)
	s.Require().Equal([]models.Container{}, list)

	list, err = s.repo.ListContainers(s.ctx, "new-namespace")
	s.Require().NoError(err)
	s.Require().Equal([]models.Container{
		{Name: "and-then-there-was-the-one", CreatedAt: time.Date(2024, 1, 2, 1, 2, 3, 0, time.UTC), VersionsTTL: 3600 * time.Second},
	}, list)
}

func (s *postgreSQLRepositoryTestSuite) TestContainersPagination() {
	s.tp.On("Now").Return("2024-01-02T01:02:03Z").Times(4)

	err := s.repo.CreateNamespace(s.ctx, "new-namespace")
	s.Require().NoError(err)

	total, list, err := s.repo.ListContainersByPage(s.ctx, defaultNamespace, 0, 100)
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), total)
	s.Require().Equal([]models.Container{}, list)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "test-container1", 1*time.Hour)
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "test-container2", 2*time.Hour)
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "test-container3", 3*time.Hour)
	s.Require().NoError(err)

	total, list, err = s.repo.ListContainersByPage(s.ctx, defaultNamespace, 0, 2)
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), total)
	s.Require().Equal([]models.Container{
		{Name: "test-container1", CreatedAt: time.Date(2024, 1, 2, 1, 2, 3, 0, time.UTC), VersionsTTL: 1 * time.Hour},
		{Name: "test-container2", CreatedAt: time.Date(2024, 1, 2, 1, 2, 3, 0, time.UTC), VersionsTTL: 2 * time.Hour},
	}, list)
}
