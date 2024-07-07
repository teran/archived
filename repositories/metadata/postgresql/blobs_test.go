package postgresql

func (s *postgreSQLRepositoryTestSuite) TestBlobs() {
	err := s.repo.CreateBLOB(s.ctx, "deadbeef", 15, "text/plain")
	s.Require().NoError(err)
}
