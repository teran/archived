package postgresql

func (s *postgreSQLRepositoryTestSuite) TestBlobs() {
	const (
		containerName = "test-container"
		checksum      = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	)

	s.tp.On("Now").Return("2024-01-02T01:02:03Z").Once()

	err := s.repo.CreateContainer(s.ctx, containerName)
	s.Require().NoError(err)

	versionID, err := s.repo.CreateVersion(s.ctx, containerName)
	s.Require().NoError(err)

	err = s.repo.CreateBLOB(s.ctx, checksum, 15, "text/plain")
	s.Require().NoError(err)

	err = s.repo.CreateObject(s.ctx, containerName, versionID, "test-object.txt", checksum)
	s.Require().NoError(err)

	err = s.repo.MarkVersionPublished(s.ctx, containerName, versionID)
	s.Require().NoError(err)

	casKey, err := s.repo.GetBlobKeyByObject(s.ctx, containerName, versionID, "test-object.txt")
	s.Require().NoError(err)
	s.Require().Equal(checksum, casKey)
}
