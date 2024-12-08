package service

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	mockRepo "github.com/teran/archived/repositories/metadata/mock"
)

func TestConfigValidate(t *testing.T) {
	type testCase struct {
		name   string
		in     *Config
		expOut error
	}

	tcs := []testCase{
		{
			name: "valid config",
			in: &Config{
				MdRepo:                   mockRepo.New(),
				DryRun:                   false,
				UnpublishedVersionMaxAge: 10 * time.Hour,
			},
		},
		{
			name: "empty config",
			in:   &Config{},
			expOut: errors.New(
				"MdRepo: cannot be blank; UnpublishedVersionMaxAge: cannot be blank.",
			),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			err := tc.in.Validate()
			if tc.expOut != nil {
				r.Error(err)
				r.Equal(tc.expOut.Error(), err.Error())
			} else {
				r.NoError(err)
			}
		})
	}
}
