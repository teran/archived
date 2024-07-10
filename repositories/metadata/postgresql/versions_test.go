package postgresql

import "github.com/teran/archived/repositories/metadata"

func (s *postgreSQLRepositoryTestSuite) TestVersionsOperations() {
	err := s.repo.CreateContainer(s.ctx, "container1")
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, "container2")
	s.Require().NoError(err)

	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Once()
	s.tp.On("Now").Return("2024-07-07T11:12:13Z").Once()

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
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Once()

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
	_, err := s.repo.ListObjects(s.ctx, "test-container", "2024-01-02T03:04:05Z", 0, 100)
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)
}

func (s *postgreSQLRepositoryTestSuite) TestListObjectsErrorsNotExistentVersion() {
	err := s.repo.CreateContainer(s.ctx, "test-container")
	s.Require().NoError(err)

	_, err = s.repo.ListObjects(s.ctx, "test-container", "2024-01-02T03:04:05Z", 0, 100)
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)
}
