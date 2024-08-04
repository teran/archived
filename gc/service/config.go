package service

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/teran/archived/repositories/metadata"
)

type Config struct {
	MdRepo                   metadata.Repository
	DryRun                   bool
	UnpublishedVersionMaxAge time.Duration
	TimeNowFunc              func() time.Time
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.MdRepo, validation.Required),
		validation.Field(&c.DryRun),
		validation.Field(&c.UnpublishedVersionMaxAge, validation.Required, validation.Min(time.Hour)),
		validation.Field(&c.TimeNowFunc, validation.Required),
	)
}
