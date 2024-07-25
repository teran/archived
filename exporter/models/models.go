package models

type VersionsCount struct {
	ContainerName string
	VersionsCount uint64
	IsPublished   bool
}

type ObjectsCount struct {
	ContainerName string
	VersionName   string
	IsPublished   bool
	ObjectsCount  uint64
}

type BlobsRawSizeBytes struct {
	ContainerName string
	VersionName   string
	IsPublished   bool
	SizeBytes     uint64
}

type Stats struct {
	ContainersCount     uint64
	VersionsCount       []VersionsCount
	ObjectsCount        []ObjectsCount
	BlobsCount          uint64
	BlobsRawSizeBytes   []BlobsRawSizeBytes
	BlobsTotalSizeBytes uint64
}
