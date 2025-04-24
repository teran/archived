package apt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func TestFetchMetadataPlain(t *testing.T) {
	r := require.New(t)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/Release", func(c echo.Context) error {
		return c.File("testdata/Release")
	})
	e.GET("/Release.gz", func(c echo.Context) error {
		return c.File("testdata/Release.gz")
	})
	e.GET("/Release.xz", func(c echo.Context) error {
		return c.File("testdata/Release.xz")
	})

	srv := httptest.NewServer(e)
	defer srv.Close()

	type testCase struct {
		filename       string
		expOutChecksum string
	}

	tcs := []testCase{
		{
			filename:       "Release",
			expOutChecksum: "633f532fd2c9e3defddb4851e48d5195e5908305d1885b15f606008b6d203cce",
		},
		{
			filename:       "Release.gz",
			expOutChecksum: "fa807a59f92ce07974633977b49b861f887280e46b396c86bb1ea01d1a2cbc18",
		},
		{
			filename:       "Release.xz",
			expOutChecksum: "2135d3d555c09687d1ab5ad985a46213511181d603680637798e89c4c843920c",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.filename, func(t *testing.T) {
			v := &ComponentRelease{}
			out, err := fetchMetadata(context.TODO(), srv.URL+"/"+tc.filename, v)
			r.NoError(err)

			checksum, err := sha256FromBytes(out)
			r.NoError(err)
			r.Equal(tc.expOutChecksum, checksum)
		})
	}
}

func TestDetectMimeTypeByFilename(t *testing.T) {
	type testCase struct {
		in     string
		expOut string
	}

	tcs := []testCase{
		{
			in:     "file.gz",
			expOut: "application/x-gzip",
		},
		{
			in:     "file.xz",
			expOut: "application/x-xz",
		},
		{
			in:     "file.deb",
			expOut: "application/vnd.debian.binary-package",
		},
		{
			in:     "file.rpm",
			expOut: "application/octet-stream",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.in, func(t *testing.T) {
			r := require.New(t)

			out := detectMimeTypeByFilename(tc.in)
			r.Equal(tc.expOut, out)
		})
	}
}

func TestSha256FromBytes(t *testing.T) {
	type testCase struct {
		in     string
		expOut string
	}

	tcs := []testCase{
		{
			in:     "test string",
			expOut: "d5579c46dfcc7f18207013e65b44e4cb4e2c2298f4ac457ba8f82743f31e930b",
		},
		{
			in:     "another string",
			expOut: "81e7826a5821395470e5a2fed0277b6a40c26257512319875e1d70106dcb1ca0",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.in, func(t *testing.T) {
			r := require.New(t)

			out, err := sha256FromBytes([]byte(tc.in))
			r.NoError(err)
			r.Equal(tc.expOut, out)
		})
	}
}

func TestGetFle(t *testing.T) {
	r := require.New(t)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/file", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "application/octet-stream", []byte("test file"))
	})

	srv := httptest.NewServer(e)
	defer srv.Close()

	out, err := getFile(context.TODO(), srv.URL+"/file")
	r.NoError(err)
	r.Equal("test file", string(out))
}
