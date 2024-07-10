package postgresql

import "github.com/teran/archived/repositories/metadata"

func (s *postgreSQLRepositoryTestSuite) TestObjects() {
	const containerName = "test-container-1"

	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Once()

	err := s.repo.CreateContainer(s.ctx, containerName)
	s.Require().NoError(err)

	versionID, err := s.repo.CreateVersion(s.ctx, containerName)
	s.Require().NoError(err)

	err = s.repo.CreateBLOB(s.ctx, "deadbeef", 10, "text/plain")
	s.Require().NoError(err)

	err = s.repo.CreateObject(s.ctx, containerName, versionID, "data/some-key.txt", "deadbeef")
	s.Require().NoError(err)

	objects, err := s.repo.ListObjects(s.ctx, containerName, versionID, 0, 100)
	s.Require().NoError(err)
	s.Require().Equal([]string{"data/some-key.txt"}, objects)

	err = s.repo.CreateBLOB(s.ctx, "deadbeef2", 10, "text/plain")
	s.Require().NoError(err)

	err = s.repo.RemapObject(s.ctx, containerName, versionID, "data/some-key.txt", "deadbeef2")
	s.Require().NoError(err)

	err = s.repo.DeleteObject(s.ctx, containerName, versionID, "data/some-key.txt")
	s.Require().NoError(err)

	objects, err = s.repo.ListObjects(s.ctx, containerName, versionID, 0, 100)
	s.Require().NoError(err)
	s.Require().Equal([]string{}, objects)
}

func (s *postgreSQLRepositoryTestSuite) TestListObjectsErrors() {
	// Nothing exists: container and version
	_, err := s.repo.ListObjects(s.ctx, "container", "version", 0, 1000)
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)

	// version doesn't exist
	err = s.repo.CreateContainer(s.ctx, "container")
	s.Require().NoError(err)

	_, err = s.repo.ListObjects(s.ctx, "container", "version", 0, 1000)
	s.Require().Error(err)
	s.Require().Equal(metadata.ErrNotFound, err)
}
