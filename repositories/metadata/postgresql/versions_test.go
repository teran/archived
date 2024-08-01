package postgresql

import "github.com/teran/archived/repositories/metadata"

func (s *postgreSQLRepositoryTestSuite) TestVersionsOperations() {
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Times(4)
	s.tp.On("Now").Return("2024-07-07T11:12:13Z").Times(2)

	err := s.repo.CreateContainer(s.ctx, "container1")
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, "container2")
	s.Require().NoError(err)

	vName, err := s.repo.CreateVersion(s.ctx, "container1")
	s.Require().NoError(err)
	s.Require().Equal("20240707101112", vName)

	vName, err = s.repo.CreateVersion(s.ctx, "container2")
	s.Require().NoError(err)
	s.Require().Equal("20240707111213", vName)

	listC1, err := s.repo.ListAllVersionsByContainer(s.ctx, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]string{"20240707101112"}, listC1)

	listC2, err := s.repo.ListAllVersionsByContainer(s.ctx, "container2")
	s.Require().NoError(err)
	s.Require().Equal([]string{"20240707111213"}, listC2)
}

func (s *postgreSQLRepositoryTestSuite) TestPublishVersion() {
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Times(3)

	err := s.repo.CreateContainer(s.ctx, "container1")
	s.Require().NoError(err)

	version, err := s.repo.CreateVersion(s.ctx, "container1")
	s.Require().NoError(err)

	list, err := s.repo.ListAllVersionsByContainer(s.ctx, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]string{"20240707101112"}, list)

	list, err = s.repo.ListPublishedVersionsByContainer(s.ctx, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]string{}, list)

	err = s.repo.MarkVersionPublished(s.ctx, "container1", version)
	s.Require().NoError(err)

	list, err = s.repo.ListAllVersionsByContainer(s.ctx, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]string{"20240707101112"}, list)

	list, err = s.repo.ListPublishedVersionsByContainer(s.ctx, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]string{"20240707101112"}, list)
}

func (s *postgreSQLRepositoryTestSuite) TestListAllVersionsByContainerErrors() {
	_, err := s.repo.ListAllVersionsByContainer(s.ctx, "test-container")
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)
}

func (s *postgreSQLRepositoryTestSuite) TestListObjectsErrorsNotExistentContainer() {
	_, _, err := s.repo.ListObjects(s.ctx, "test-container", "2024-01-02T03:04:05Z", 0, 100)
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)
}

func (s *postgreSQLRepositoryTestSuite) TestListObjectsErrorsNotExistentVersion() {
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Once()

	err := s.repo.CreateContainer(s.ctx, "test-container")
	s.Require().NoError(err)

	_, _, err = s.repo.ListObjects(s.ctx, "test-container", "2024-01-02T03:04:05Z", 0, 100)
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)
}

func (s *postgreSQLRepositoryTestSuite) TestVersionsPagination() {
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Times(3)
	s.tp.On("Now").Return("2024-07-07T10:11:13Z").Times(2)
	s.tp.On("Now").Return("2024-07-07T10:11:14Z").Times(2)

	err := s.repo.CreateContainer(s.ctx, "container1")
	s.Require().NoError(err)

	version1, err := s.repo.CreateVersion(s.ctx, "container1")
	s.Require().NoError(err)

	err = s.repo.MarkVersionPublished(s.ctx, "container1", version1)
	s.Require().NoError(err)

	_, err = s.repo.CreateVersion(s.ctx, "container1")
	s.Require().NoError(err)

	version3, err := s.repo.CreateVersion(s.ctx, "container1")
	s.Require().NoError(err)

	err = s.repo.MarkVersionPublished(s.ctx, "container1", version3)
	s.Require().NoError(err)

	total, listByPage, err := s.repo.ListPublishedVersionsByContainerAndPage(s.ctx, "container1", 0, 5)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), total)
	s.Require().Equal([]string{version1, version3}, listByPage)

	total, listByPage, err = s.repo.ListPublishedVersionsByContainerAndPage(s.ctx, "container1", 1, 2)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), total)
	s.Require().Equal([]string{version3}, listByPage)
}

func (s *postgreSQLRepositoryTestSuite) TestDeleteVersion() {
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Times(4)
	s.tp.On("Now").Return("2024-07-07T10:11:13Z").Times(2)
	s.tp.On("Now").Return("2024-07-07T10:11:14Z").Times(2)
	s.tp.On("Now").Return("2024-07-07T10:11:15Z").Times(2)

	err := s.repo.CreateContainer(s.ctx, "container1")
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, "container2")
	s.Require().NoError(err)

	version1, err := s.repo.CreateVersion(s.ctx, "container1")
	s.Require().NoError(err)

	version2, err := s.repo.CreateVersion(s.ctx, "container1")
	s.Require().NoError(err)

	version3, err := s.repo.CreateVersion(s.ctx, "container2")
	s.Require().NoError(err)

	version4, err := s.repo.CreateVersion(s.ctx, "container2")
	s.Require().NoError(err)

	versions1, err := s.repo.ListAllVersionsByContainer(s.ctx, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]string{version1, version2}, versions1)

	versions2, err := s.repo.ListAllVersionsByContainer(s.ctx, "container2")
	s.Require().NoError(err)
	s.Require().Equal([]string{version3, version4}, versions2)

	err = s.repo.DeleteVersion(s.ctx, "container1", version1)
	s.Require().NoError(err)

	versions1, err = s.repo.ListAllVersionsByContainer(s.ctx, "container1")
	s.Require().NoError(err)
	s.Require().Equal([]string{version2}, versions1)

	versions2, err = s.repo.ListAllVersionsByContainer(s.ctx, "container2")
	s.Require().NoError(err)
	s.Require().Equal([]string{version3, version4}, versions2)
}

func (s *postgreSQLRepositoryTestSuite) TestGetLatestPublishedVersionByContainer() {
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Times(4)
	s.tp.On("Now").Return("2024-07-07T10:11:13Z").Times(2)
	s.tp.On("Now").Return("2024-07-07T10:11:14Z").Times(2)
	s.tp.On("Now").Return("2024-07-07T10:11:15Z").Times(2)
	s.tp.On("Now").Return("2024-07-07T10:11:16Z").Times(2)

	err := s.repo.CreateContainer(s.ctx, "container1")
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, "container2")
	s.Require().NoError(err)

	_, err = s.repo.CreateVersion(s.ctx, "container1")
	s.Require().NoError(err)

	version2, err := s.repo.CreateVersion(s.ctx, "container1")
	s.Require().NoError(err)

	err = s.repo.MarkVersionPublished(s.ctx, "container1", version2)
	s.Require().NoError(err)

	_, err = s.repo.CreateVersion(s.ctx, "container1")
	s.Require().NoError(err)

	_, err = s.repo.CreateVersion(s.ctx, "container2")
	s.Require().NoError(err)

	_, err = s.repo.CreateVersion(s.ctx, "container2")
	s.Require().NoError(err)

	versionName, err := s.repo.GetLatestPublishedVersionByContainer(s.ctx, "container1")
	s.Require().NoError(err)
	s.Require().Equal(version2, versionName)

	_, err = s.repo.GetLatestPublishedVersionByContainer(s.ctx, "container2")
	s.Require().Error(err)
	s.Require().Equal("not found", err.Error())
}
