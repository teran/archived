package postgresql

import "github.com/teran/archived/repositories/metadata"

func (s *postgreSQLRepositoryTestSuite) TestNamespaceOperations() {
	s.tp.On("Now").Return("2024-01-02T01:02:03Z").Times(3)

	list, err := s.repo.ListNamespaces(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{"default"}, list)

	err = s.repo.CreateNamespace(s.ctx, "test-namespace9")
	s.Require().NoError(err)

	err = s.repo.CreateNamespace(s.ctx, "test-namespace5")
	s.Require().NoError(err)

	err = s.repo.CreateNamespace(s.ctx, "test-namespace9")
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrConflict, err)

	list, err = s.repo.ListNamespaces(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"default",
		"test-namespace5",
		"test-namespace9",
	}, list)

	err = s.repo.DeleteNamespace(s.ctx, "test-namespace9")
	s.Require().NoError(err)

	err = s.repo.DeleteNamespace(s.ctx, "not-existent")
	s.Require().NoError(err)

	list, err = s.repo.ListNamespaces(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"default",
		"test-namespace5",
	}, list)

	err = s.repo.RenameNamespace(s.ctx, "test-namespace5", "and-then-there-was-the-one")
	s.Require().NoError(err)

	err = s.repo.RenameNamespace(s.ctx, "not-existent", "some-name")
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)

	list, err = s.repo.ListNamespaces(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"and-then-there-was-the-one",
		"default",
	}, list)
}
