package lazyblob

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

func TestLazyblob(t *testing.T) {
	ctx := context.TODO()
	r := require.New(t)

	m := &testHandlerMock{}
	defer m.AssertExpectations(t)

	m.On("StaticFile", "/first").Return(http.StatusOK, "text/plain", []byte("test data")).Once()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/*", m.StaticFile)

	srv := httptest.NewServer(e)
	defer srv.Close()

	lb := New(srv.URL+"/first", t.TempDir(), 9)
	defer func() { _ = lb.Close() }()

	fn, err := lb.Filename(ctx)
	r.NoError(err)
	r.True(strings.HasSuffix(fn, ".tmp"))

	fp, err := lb.Reader(ctx)
	r.NoError(err)

	data, err := io.ReadAll(fp)
	r.NoError(err)
	r.Equal("test data", string(data))

	url := lb.URL()
	r.Equal(srv.URL+"/first", url)
}

func TestLazyblobNotFound(t *testing.T) {
	ctx := context.TODO()
	r := require.New(t)

	m := &testHandlerMock{}
	defer m.AssertExpectations(t)

	m.On("StaticFile", "/not-found").Return(http.StatusNotFound, "text/plain", []byte("not found")).Twice()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/*", m.StaticFile)

	srv := httptest.NewServer(e)
	defer srv.Close()

	lb := New(srv.URL+"/not-found", t.TempDir(), 9)
	defer func() { _ = lb.Close() }()

	_, err := lb.Filename(ctx)
	r.Error(err)
	r.Equal(srv.URL+"/not-found: unexpected HTTP response status: 404 Not Found", err.Error())

	_, err = lb.Reader(ctx)
	r.Error(err)
	r.Equal("error downloading file: "+srv.URL+"/not-found: unexpected HTTP response status: 404 Not Found", err.Error())

	url := lb.URL()
	r.Equal(srv.URL+"/not-found", url)
}

type testHandlerMock struct {
	mock.Mock
}

func (m *testHandlerMock) StaticFile(c echo.Context) error {
	args := m.Called(c.Request().RequestURI)
	return c.Blob(args.Int(0), args.String(1), args.Get(2).([]byte))
}
