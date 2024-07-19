package postgresql

func (s *postgreSQLRepositoryTestSuite) TestContainerOperations() {
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
		`error executing SQL query: pq: duplicate key value violates unique constraint "containers_name_key"`,
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
}
