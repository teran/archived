package models

type VersionsCount struct {
	Namespace     string
	ContainerName string
	VersionsCount uint64
	IsPublished   bool
}

type ObjectsCount struct {
	Namespace     string
	ContainerName string
	VersionName   string
	IsPublished   bool
	ObjectsCount  uint64
}

type BlobsRawSizeBytes struct {
	Namespace     string
	ContainerName string
	VersionName   string
	IsPublished   bool
	SizeBytes     uint64
}

type Stats struct {
	NamespacesCount     uint64
	ContainersCount     uint64
	VersionsCount       []VersionsCount
	ObjectsCount        []ObjectsCount
	BlobsCount          uint64
	BlobsRawSizeBytes   []BlobsRawSizeBytes
	BlobsTotalSizeBytes uint64
}
