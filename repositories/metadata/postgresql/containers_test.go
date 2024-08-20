package postgresql

func (s *postgreSQLRepositoryTestSuite) TestContainerOperations() {
	s.tp.On("Now").Return("2024-01-02T01:02:03Z").Times(3)

	list, err := s.repo.ListContainers(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{}, list)

	err = s.repo.CreateContainer(s.ctx, "test-container9")
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, "test-container5")
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, "test-container9")
	s.Require().Error(err)
	s.Require().Equal(
		`pq: duplicate key value violates unique constraint "containers_name_key"`,
		err.Error(),
	)

	list, err = s.repo.ListContainers(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"test-container5",
		"test-container9",
	}, list)

	err = s.repo.DeleteContainer(s.ctx, "test-container9")
	s.Require().NoError(err)

	list, err = s.repo.ListContainers(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"test-container5",
	}, list)

	err = s.repo.RenameContainer(s.ctx, "test-container5", "and-then-there-was-the-one")
	s.Require().NoError(err)

	list, err = s.repo.ListContainers(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"and-then-there-was-the-one",
	}, list)
}
