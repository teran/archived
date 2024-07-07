package postgresql

func (s *postgreSQLRepositoryTestSuite) TestContainerOperations() {
	list, err := s.repo.ListContainers(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{}, list)

	err = s.repo.CreateContainer(s.ctx, "test-container")
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, "test-container2")
	s.Require().NoError(err)

	err = s.repo.CreateContainer(s.ctx, "test-container")
	s.Require().Error(err)
	s.Require().Equal(
		`error executing SQL query: pq: duplicate key value violates unique constraint "containers_name_key"`,
		err.Error(),
	)

	list, err = s.repo.ListContainers(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"test-container",
		"test-container2",
	}, list)

	err = s.repo.DeleteContainer(s.ctx, "test-container")
	s.Require().NoError(err)

	list, err = s.repo.ListContainers(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]string{
		"test-container2",
	}, list)
}
