package postgresql

import "github.com/teran/archived/repositories/metadata"

func (s *postgreSQLRepositoryTestSuite) TestContainerOperations() {
	s.tp.On("Now").Return("2024-01-02T01:02:03Z").Times(3)

	list, err := s.repo.ListContainers(s.ctx, defaultNamespace)
	s.Require().NoError(err)
	s.Require().Equal([]string{}, list)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "test-container9")
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "test-container5")
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, defaultNamespace, "test-container9")
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrConflict, err)

	list, err = s.repo.ListContainers(s.ctx, defaultNamespace)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"test-container5",
		"test-container9",
	}, list)

	err = s.repo.DeleteContainer(s.ctx, defaultNamespace, "test-container9")
	s.Require().NoError(err)

	err = s.repo.DeleteContainer(s.ctx, defaultNamespace, "not-existent")
	s.Require().NoError(err)

	list, err = s.repo.ListContainers(s.ctx, defaultNamespace)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"test-container5",
	}, list)

	err = s.repo.RenameContainer(s.ctx, defaultNamespace, "test-container5", "and-then-there-was-the-one")
	s.Require().NoError(err)

	err = s.repo.RenameContainer(s.ctx, defaultNamespace, "not-existent", "some-name")
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)

	list, err = s.repo.ListContainers(s.ctx, defaultNamespace)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"and-then-there-was-the-one",
	}, list)
}
