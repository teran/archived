package models

import "time"

type Container struct {
	Name        string
	CreatedAt   time.Time
	VersionsTTL time.Duration
}
