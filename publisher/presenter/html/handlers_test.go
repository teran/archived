package html

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/suite"
	"github.com/teran/archived/models"
	"github.com/teran/archived/service"
)

const defaultNamespace = "default"

func (s *handlersTestSuite) TestNamespaceIndex() {
	s.serviceMock.On("ListNamespaces").Return([]string{"default", "test-namespace-1"}, nil).Once()

	s.compareHTMLResponse(s.srv.URL, "testdata/namespaces.html.sample")
}

func (s *handlersTestSuite) TestContainerIndex() {
	s.serviceMock.On("ListContainersByPage", defaultNamespace, uint64(1)).Return(uint64(100), []models.Container{{Name: "test-container-1"}}, nil).Once()

	s.compareHTMLResponse(s.srv.URL+"/default/", "testdata/containers.html.sample")
}

func (s *handlersTestSuite) TestVersionIndex() {
	s.serviceMock.On("ListPublishedVersionsByPage", defaultNamespace, "test-container-1", uint64(1)).Return(uint64(100), []models.Version{
		{Name: "20241011121314"},
	}, nil).Once()

	s.compareHTMLResponse(s.srv.URL+"/default/test-container-1/", "testdata/versions.html.sample")
}

func (s *handlersTestSuite) TestObjectIndex() {
	s.serviceMock.On("ListObjectsByPage", defaultNamespace, "test-container-1", "20241011121314", uint64(1)).Return(uint64(100), []string{"test-object-dir/file.txt"}, nil).Once()

	s.compareHTMLResponse(s.srv.URL+"/default/test-container-1/20241011121314/", "testdata/objects.html.sample")
}

func (s *handlersTestSuite) TestGetObject() {
	s.serviceMock.On("GetObjectURL", defaultNamespace, "test-container-1", "20241011121314", "test-dir/filename.txt").Return("https://example.com/some-addr", nil).Once()

	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, s.srv.URL+"/default/test-container-1/20241011121314/test-dir/filename.txt", nil)
	s.Require().NoError(err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	v := resp.Header.Get("Location")
	s.Require().Equal("https://example.com/some-addr", v)
}

func (s *handlersTestSuite) TestGetObjectSchemeMismatchXForwardedScheme() {
	s.serviceMock.On("GetObjectURL", defaultNamespace, "test-container-1", "20241011121314", "test-dir/filename.txt").Return("https://example.com/some-addr", nil).Once()

	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, s.srv.URL+"/default/test-container-1/20241011121314/test-dir/filename.txt", nil)
	s.Require().NoError(err)

	req.Header.Set("X-Forwarded-Scheme", "http")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	v := resp.Header.Get("Location")
	s.Require().Equal("http://example.com/some-addr", v)
}

func (s *handlersTestSuite) TestGetObjectSchemeMismatchXScheme() {
	s.serviceMock.On("GetObjectURL", defaultNamespace, "test-container-1", "20241011121314", "test-dir/filename.txt").Return("https://example.com/some-addr", nil).Once()

	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, s.srv.URL+"/default/test-container-1/20241011121314/test-dir/filename.txt", nil)
	s.Require().NoError(err)

	req.Header.Set("X-Scheme", "http")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	v := resp.Header.Get("Location")
	s.Require().Equal("http://example.com/some-addr", v)
}

func (s *handlersTestSuite) TestErrNotFound() {
	s.serviceMock.On("ListContainersByPage", defaultNamespace, uint64(1)).Return(uint64(100), []models.Container(nil), service.ErrNotFound).Once()
	s.compareHTMLResponse(s.srv.URL+"/default/", "testdata/404.html.sample")

	s.serviceMock.On("ListPublishedVersionsByPage", defaultNamespace, "test-container-1", uint64(1)).Return(uint64(100), []models.Version(nil), service.ErrNotFound).Once()
	s.compareHTMLResponse(s.srv.URL+"/default/test-container-1/", "testdata/404.html.sample")

	s.serviceMock.On("ListObjectsByPage", defaultNamespace, "test-container-1", "20240101010101", uint64(1)).Return(uint64(100), []string(nil), service.ErrNotFound).Once()
	s.compareHTMLResponse(s.srv.URL+"/default/test-container-1/20240101010101/", "testdata/404.html.sample")

	s.serviceMock.On("GetObjectURL", defaultNamespace, "test-container-1", "20240101010101", "test-object.txt").Return("", service.ErrNotFound).Once()
	s.compareHTMLResponse(s.srv.URL+"/default/test-container-1/20240101010101/test-object.txt", "testdata/404.html.sample")

	s.compareHTMLResponse(s.srv.URL+"/default/test-container-1/20240101010101", "testdata/404.html.sample")
	s.compareHTMLResponse(s.srv.URL+"/default/test-container-1", "testdata/404.html.sample")
}

func (s *handlersTestSuite) TestErr5xx() {
	s.serviceMock.On("ListContainersByPage", defaultNamespace, uint64(1)).Panic("blah").Once()
	s.compareHTMLResponse(s.srv.URL+"/default/", "testdata/5xx.html.sample")

	s.serviceMock.On("ListPublishedVersionsByPage", defaultNamespace, "test-container-1", uint64(1)).Panic("blah").Once()
	s.compareHTMLResponse(s.srv.URL+"/default/test-container-1/", "testdata/5xx.html.sample")
}

func (s *handlersTestSuite) TestEscapedPath() {
	s.serviceMock.On("GetObjectURL", defaultNamespace, "test-container-1", "20240101010101", "test object.txt").Return("https://example.com/some-addr", nil).Once()

	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, s.srv.URL+"/default/test-container-1/20240101010101/test%20object.txt", nil)
	s.Require().NoError(err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	v := resp.Header.Get("Location")
	s.Require().Equal("https://example.com/some-addr", v)
}

// Definitions ...
type handlersTestSuite struct {
	suite.Suite

	ctx context.Context

	srv *httptest.Server

	serviceMock *service.Mock
	handlers    Handlers
}

func (s *handlersTestSuite) SetupTest() {
	s.ctx = context.TODO()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	s.serviceMock = service.NewMock()

	s.handlers = New(s.serviceMock, "templates", "static", true)
	s.handlers.Register(e)

	s.srv = httptest.NewServer(e)
}

func (s *handlersTestSuite) TearDownTest() {
	s.serviceMock.AssertExpectations(s.T())

	s.srv.Close()
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, &handlersTestSuite{})
}

func (s *handlersTestSuite) compareHTMLResponse(url, responseSamplePath string) {
	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, url, nil)
	s.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	sampleData, err := os.ReadFile(responseSamplePath)
	s.Require().NoError(err)

	data, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	s.Require().Equal(string(sampleData), string(data))
}
