package html

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/suite"
	"github.com/teran/archived/service"
)

func (s *handlersTestSuite) TestContainerIndex() {
	s.serviceMock.On("ListContainers").Return([]string{"test-container-1"}, nil).Once()

	s.compareHTMLResponse(s.srv.URL, "testdata/index.html.sample")
}

func (s *handlersTestSuite) TestVersionIndex() {
	s.serviceMock.On("ListVersions", "test-container-1").Return([]string{"20241011121314"}, nil).Once()

	s.compareHTMLResponse(s.srv.URL+"/test-container-1/", "testdata/versions.html.sample")
}

func (s *handlersTestSuite) TestObjectIndex() {
	s.serviceMock.On("ListObjects", "test-container-1", "20241011121314").Return([]string{"test-object-dir/file.txt"}, nil).Once()

	s.compareHTMLResponse(s.srv.URL+"/test-container-1/20241011121314/", "testdata/objects.html.sample")
}

func (s *handlersTestSuite) TestGetObject() {
	s.serviceMock.On("GetObjectURL", "test-container-1", "20241011121314", "test-dir/filename.txt").Return("https://example.com/some-addr", nil).Once()

	req, err := http.NewRequest(http.MethodGet, s.srv.URL+"/test-container-1/20241011121314/test-dir/filename.txt", nil)
	s.Require().NoError(err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	s.Require().NoError(err)

	v := resp.Header.Get("Location")
	s.Require().Equal("https://example.com/some-addr", v)
}

func (s *handlersTestSuite) TestErrNotFound() {
	s.serviceMock.On("ListVersions", "test-container-1").Return([]string(nil), service.ErrNotFound).Once()
	s.compareHTMLResponse(s.srv.URL+"/test-container-1/", "testdata/404.html.sample")

	s.serviceMock.On("ListObjects", "test-container-1", "20240101010101").Return([]string(nil), service.ErrNotFound).Once()
	s.compareHTMLResponse(s.srv.URL+"/test-container-1/20240101010101/", "testdata/404.html.sample")

	s.serviceMock.On("GetObjectURL", "test-container-1", "20240101010101", "test-object.txt").Return("", service.ErrNotFound).Once()
	s.compareHTMLResponse(s.srv.URL+"/test-container-1/20240101010101/test-object.txt", "testdata/404.html.sample")

	s.compareHTMLResponse(s.srv.URL+"/test-container-1/20240101010101", "testdata/404.html.sample")
	s.compareHTMLResponse(s.srv.URL+"/test-container-1", "testdata/404.html.sample")
}

func (s *handlersTestSuite) TestErr5xx() {
	s.serviceMock.On("ListVersions", "test-container-1").Panic("blah").Once()
	s.compareHTMLResponse(s.srv.URL+"/test-container-1/", "testdata/5xx.html.sample")
}

// Definitions ...
type handlersTestSuite struct {
	suite.Suite

	srv *httptest.Server

	serviceMock *service.Mock
	handlers    Handlers
}

func (s *handlersTestSuite) SetupTest() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	s.serviceMock = service.NewMock()

	s.handlers = New(s.serviceMock, "templates")
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
	req, err := http.NewRequest(http.MethodGet, url, nil)
	s.Require().NoError(err)

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)

	sampleData, err := os.ReadFile(responseSamplePath)
	s.Require().NoError(err)

	data, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	s.Require().Equal(sampleData, data)
}
