package mirrorlist

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func TestGetMirrors(t *testing.T) {
	r := require.New(t)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/mirrorlist", mirrorlistHandler)

	srv := httptest.NewServer(e)
	defer srv.Close()

	ml, err := newWithRandom(context.TODO(), srv.URL+"/mirrorlist", testRandomNumber)
	r.NoError(err)
	r.Equal([]string{
		"https://mirror.tzulo.com/rocky/9.4/AppStream/x86_64/os/",
		"http://mirror.cs.vt.edu/pub/rocky/9.4/AppStream/x86_64/os/",
		"http://mirrors.rit.edu/rocky/9.4/AppStream/x86_64/os/",
		"http://ash.mirrors.clouvider.net/rocky/9.4/AppStream/x86_64/os",
	}, ml.(*mirrorlist).mirrors)

	r.Equal("https://mirror.tzulo.com/rocky/9.4/AppStream/x86_64/os/", ml.URL(SelectModeFirstOnly))
	r.Equal("http://mirrors.rit.edu/rocky/9.4/AppStream/x86_64/os/", ml.URL(SelectModeRandom))
	r.Equal("https://mirror.tzulo.com/rocky/9.4/AppStream/x86_64/os/", ml.URL(SelectMode("blah")))
}

func TestEmptyMirrorlist(t *testing.T) {
	r := require.New(t)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/empty", emptyMirrorlistHandler)

	srv := httptest.NewServer(e)
	defer srv.Close()

	_, err := newWithRandom(context.TODO(), srv.URL+"/mirrorlist", testRandomNumber)
	r.Error(err)
	r.Equal(ErrEmptyMirrorlist, err)
}

func mirrorlistHandler(c echo.Context) error {
	// Sample data received as a first few lines of
	// 	https://mirrors.rockylinux.org/mirrorlist?arch=x86_64&repo=rocky-AppStream-9.4&country=US
	return c.Blob(http.StatusOK, "text/plain", []byte(strings.Join([]string{
		"# repo = rocky-AppStream-9.4 arch = x86_64 country = US",
		"https://mirror.tzulo.com/rocky/9.4/AppStream/x86_64/os/",
		"http://mirror.cs.vt.edu/pub/rocky/9.4/AppStream/x86_64/os/",
		"http://mirrors.rit.edu/rocky/9.4/AppStream/x86_64/os/",
		"http://ash.mirrors.clouvider.net/rocky/9.4/AppStream/x86_64/os",
		"some text",
	}, "\n")))
}

func emptyMirrorlistHandler(c echo.Context) error {
	return c.Blob(http.StatusOK, "text/plain", []byte(strings.Join([]string{
		"# repo = rocky-AppStream-9.4 arch = x86_64 country = US",
		"some text",
	}, "\n")))
}

func testRandomNumber(int) int {
	// Strong random number for test purposes
	return 2
}
