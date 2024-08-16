package yum

import (
	"context"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/teran/archived/cli/yum/models"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func TestPackages(t *testing.T) {
	r := require.New(t)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "testdata/repo")

	srv := httptest.NewServer(e)
	defer srv.Close()

	repo := New(srv.URL)

	packages, err := repo.Packages(context.Background())
	r.NoError(err)
	r.Equal([]models.Package{
		{
			Name:         "Packages/testpkg-1-1.src.rpm",
			Checksum:     "684303227d799ffe1f0b39e030a12ad249931a11ec1690e2079f981cc16d8c52",
			ChecksumType: "sha256",
			Size:         6156,
		},
		{
			Name:         "Packages/testpkg-1-1.x86_64.rpm",
			Checksum:     "d9ae5e56ea38d2ac470f320cade63663dae6ab8b8e1630b2fd5a3c607f45e2ee",
			ChecksumType: "sha256",
			Size:         6722,
		},
	}, packages)
}

func TestMetadataSHA256(t *testing.T) {
	r := require.New(t)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "testdata/repo")

	srv := httptest.NewServer(e)
	defer srv.Close()

	repo := New(srv.URL)

	_, err := repo.Packages(context.Background())
	r.NoError(err)

	md := repo.Metadata()
	r.Equal(map[string]int{
		"repodata/1b4aca205bffe8d65f33b066e3f9965cb4c009e3c94b3f296cce8bff166ad8ed-primary.sqlite.bz2":   1995,
		"repodata/2267234d92017b049818be743f720f37c176a3b3bb3e802ee4d5cd0090651091-primary.xml.gz":       720,
		"repodata/2623c0a1472f574989dcba85417e8ce27b87983bba12922a6d91d574e617d2f6-filelists.sqlite.bz2": 858,
		"repodata/314e73564000b8a68848551ce0fa9b36e11ed609698f232fa9ab5810ec531de1-filelists.xml.gz":     313,
		"repodata/64f4875d92a3672f62a2d15d5f0ae6f0806451f42403bd07105214e1c9f4f0d7-other.sqlite.bz2":     749,
		"repodata/e3984def0f3b5ce1b174fad2f6eb3c05829633d2d5d5d8ba05c9720ad59046e7-other.xml.gz":         281,
		"repodata/repomd.xml": 3069,
	}, func() map[string]int {
		keys := map[string]int{}
		for k, v := range md {
			keys[k] = len(v)
		}
		return keys
	}())
}

func TestMetadataSHA1(t *testing.T) {
	r := require.New(t)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "testdata/repo-sha1")

	srv := httptest.NewServer(e)
	defer srv.Close()

	repo := New(srv.URL)

	_, err := repo.Packages(context.Background())
	r.NoError(err)

	md := repo.Metadata()
	r.Equal(map[string]int{
		"repodata/repomd.xml": 2601,
		"repodata/4a11e3eeb25d21b08f41e5578d702d2bea21a2e7-filelists.xml.gz":     282,
		"repodata/fdedb6ce109127d52228d01b0239010ddca14c8f-other.xml.gz":         247,
		"repodata/e7a8a53e7398f6c22894718ea227fea60f2b78ba-primary.sqlite.bz2":   1937,
		"repodata/c66ce2caa41ed83879f9b3dd9f40e61c65af499e-filelists.sqlite.bz2": 787,
		"repodata/b31561a27d014d35b59b27c27859bb1c17ac573e-other.sqlite.bz2":     669,
		"repodata/80779e2ab55e25a77124d370de1d08deae8f1cc6-primary.xml.gz":       688,
	}, func() map[string]int {
		keys := map[string]int{}
		for k, v := range md {
			keys[k] = len(v)
		}
		return keys
	}())
}
