package models

import "time"

type Version struct {
	Name        string
	IsPublished bool
	CreatedAt   time.Time
}
