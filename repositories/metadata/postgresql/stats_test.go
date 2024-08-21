package postgresql

import "github.com/teran/archived/exporter/models"

func (s *postgreSQLRepositoryTestSuite) TestCountStats() {
	const containerName = "test-container-1"

	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Times(3)
	s.tp.On("Now").Return("2024-07-07T10:11:13Z").Times(7)
	s.tp.On("Now").Return("2024-07-07T10:11:14Z").Twice()

	// Create container
	err := s.repo.CreateContainer(s.ctx, defaultNamespace, containerName)
	s.Require().NoError(err)

	// Create first version
	versionID, err := s.repo.CreateVersion(s.ctx, defaultNamespace, containerName)
	s.Require().NoError(err)

	err = s.repo.CreateBLOB(s.ctx, "deadbeef", 10, "text/plain")
	s.Require().NoError(err)

	err = s.repo.CreateObject(s.ctx, defaultNamespace, containerName, versionID, "data/some-key.txt", "deadbeef")
	s.Require().NoError(err)

	err = s.repo.CreateObject(s.ctx, defaultNamespace, containerName, versionID, "data/some-key2.txt", "deadbeef")
	s.Require().NoError(err)

	// Create second version
	versionID2, err := s.repo.CreateVersion(s.ctx, defaultNamespace, containerName)
	s.Require().NoError(err)

	err = s.repo.CreateObject(s.ctx, defaultNamespace, containerName, versionID2, "data/some-key.txt", "deadbeef")
	s.Require().NoError(err)

	err = s.repo.MarkVersionPublished(s.ctx, defaultNamespace, containerName, versionID2)
	s.Require().NoError(err)

	// Count stats
	stats, err := s.repo.CountStats(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(&models.Stats{
		NamespacesCount: 1,
		ContainersCount: 1,
		VersionsCount: []models.VersionsCount{
			{
				Namespace:     defaultNamespace,
				ContainerName: "test-container-1",
				VersionsCount: 1,
				IsPublished:   false,
			},
			{
				Namespace:     defaultNamespace,
				ContainerName: "test-container-1",
				VersionsCount: 1,
				IsPublished:   true,
			},
		},
		ObjectsCount: []models.ObjectsCount{
			{
				Namespace:     defaultNamespace,
				ContainerName: "test-container-1",
				VersionName:   "20240707101112",
				IsPublished:   false,
				ObjectsCount:  2,
			},
			{
				Namespace:     defaultNamespace,
				ContainerName: "test-container-1",
				VersionName:   "20240707101113",
				IsPublished:   true,
				ObjectsCount:  1,
			},
		},
		BlobsCount: 1,
		BlobsRawSizeBytes: []models.BlobsRawSizeBytes{
			{
				Namespace:     defaultNamespace,
				ContainerName: "test-container-1",
				VersionName:   "20240707101112",
				IsPublished:   false,
				SizeBytes:     20,
			},
			{
				Namespace:     defaultNamespace,
				ContainerName: "test-container-1",
				VersionName:   "20240707101113",
				IsPublished:   true,
				SizeBytes:     10,
			},
		},
		BlobsTotalSizeBytes: 10,
	}, stats)
}
