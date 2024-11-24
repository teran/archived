package models

type Package struct {
	Name         string
	Checksum     string
	ChecksumType string
	Size         uint64
}
