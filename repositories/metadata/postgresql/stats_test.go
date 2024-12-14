package postgresql

import (
	"fmt"

	"github.com/teran/archived/exporter/models"
)

func (s *postgreSQLRepositoryTestSuite) TestCountStats() {
	const containerName = "test-container-1"

	// CreateNamespace (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Times(5)

	// CreateContainer (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:12Z").Times(10)

	// CreateVersion (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:13Z").Once()

	// CreateBLOB, 2xCreateObject (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:14Z").Times(3)

	// CreateVersion (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:15Z").Once()

	// CreateObject (created_at)
	s.tp.On("Now").Return("2024-07-07T10:11:16Z").Once()

	// Create namespaces
	for i := 1; i <= 5; i++ {
		err := s.repo.CreateNamespace(s.ctx, fmt.Sprintf("test-namespace-%d", i))
		s.Require().NoError(err)
	}

	// Create containers
	for i := 1; i <= 10; i++ {
		err := s.repo.CreateContainer(s.ctx, defaultNamespace, fmt.Sprintf("test-container-%d", i), -1)
		s.Require().NoError(err)
	}

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
		NamespacesCount: 6,
		ContainersCount: 10,
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
				VersionName:   "20240707101113",
				IsPublished:   false,
				ObjectsCount:  2,
			},
			{
				Namespace:     defaultNamespace,
				ContainerName: "test-container-1",
				VersionName:   "20240707101115",
				IsPublished:   true,
				ObjectsCount:  1,
			},
		},
		BlobsCount: 1,
		BlobsRawSizeBytes: []models.BlobsRawSizeBytes{
			{
				Namespace:     defaultNamespace,
				ContainerName: "test-container-1",
				VersionName:   "20240707101113",
				IsPublished:   false,
				SizeBytes:     20,
			},
			{
				Namespace:     defaultNamespace,
				ContainerName: "test-container-1",
				VersionName:   "20240707101115",
				IsPublished:   true,
				SizeBytes:     10,
			},
		},
		BlobsTotalSizeBytes: 10,
	}, stats)
}
