package postgresql

import "github.com/teran/archived/exporter/models"

func (s *postgreSQLRepositoryTestSuite) TestCountStats() {
	const containerName = "test-container-1"

	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Once()
	s.tp.On("Now").Return("2024-07-08T10:11:12Z").Once()

	// Create container
	err := s.repo.CreateContainer(s.ctx, containerName)
	s.Require().NoError(err)

	// Create first version
	versionID, err := s.repo.CreateVersion(s.ctx, containerName)
	s.Require().NoError(err)

	err = s.repo.CreateBLOB(s.ctx, "deadbeef", 10, "text/plain")
	s.Require().NoError(err)

	err = s.repo.CreateObject(s.ctx, containerName, versionID, "data/some-key.txt", "deadbeef")
	s.Require().NoError(err)

	err = s.repo.CreateObject(s.ctx, containerName, versionID, "data/some-key2.txt", "deadbeef")
	s.Require().NoError(err)

	// Create second version
	versionID2, err := s.repo.CreateVersion(s.ctx, containerName)
	s.Require().NoError(err)

	err = s.repo.CreateObject(s.ctx, containerName, versionID2, "data/some-key.txt", "deadbeef")
	s.Require().NoError(err)

	err = s.repo.MarkVersionPublished(s.ctx, containerName, versionID2)
	s.Require().NoError(err)

	// Count stats
	stats, err := s.repo.CountStats(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(&models.Stats{
		ContainersCount: 1,
		VersionsCount: []models.VersionsCount{
			{
				ContainerName: "test-container-1",
				VersionsCount: 1,
				IsPublished:   false,
			},
			{
				ContainerName: "test-container-1",
				VersionsCount: 1,
				IsPublished:   true,
			},
		},
		ObjectsCount: []models.ObjectsCount{
			{
				ContainerName: "test-container-1",
				VersionName:   "20240707101112",
				IsPublished:   false,
				ObjectsCount:  2,
			},
			{
				ContainerName: "test-container-1",
				VersionName:   "20240708101112",
				IsPublished:   true,
				ObjectsCount:  1,
			},
		},
		BlobsCount: 1,
		BlobsRawSizeBytes: []models.BlobsRawSizeBytes{
			{
				ContainerName: "test-container-1",
				VersionName:   "20240707101112",
				IsPublished:   false,
				SizeBytes:     20,
			},
			{
				ContainerName: "test-container-1",
				VersionName:   "20240708101112",
				IsPublished:   true,
				SizeBytes:     10,
			},
		},
		BlobsTotalSizeBytes: 10,
	}, stats)
}
