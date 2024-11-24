package yum

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChecksumFile(t *testing.T) {
	r := require.New(t)

	sha256, err := checksumFile("testdata/gpg/somekey.gpg")
	r.NoError(err)
	r.Equal("aa392a2005c38f10ce21034d6d1aaace5bbee1c3d98ac1ee06a42336d741473e", sha256)
}

func TestChecksumFileNotExistent(t *testing.T) {
	r := require.New(t)

	_, err := checksumFile("testdata/gpg/not-existent.gpg")
	r.Error(err)
	r.Equal("error performing stat on file: stat testdata/gpg/not-existent.gpg: no such file or directory", err.Error())
}
