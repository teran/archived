package html

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"strconv"

	sprig "github.com/Masterminds/sprig/v3"
	echo "github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/teran/archived/models"
	"github.com/teran/archived/service"
)

const (
	notFoundTemplateFilename    = "404.html"
	serverErrorTemplateFilename = "5xx.html"
)

type Handlers interface {
	ContainerIndex(c echo.Context) error
	VersionIndex(c echo.Context) error

	Register(e *echo.Echo)
}

type handlers struct {
	svc                      service.Publisher
	staticDir                string
	templateDir              string
	preserveSchemeOnRedirect bool
}

func New(svc service.Publisher, templateDir, staticDir string, preserveSchemeOnRedirect bool) Handlers {
	return &handlers{
		svc:                      svc,
		staticDir:                staticDir,
		templateDir:              templateDir,
		preserveSchemeOnRedirect: preserveSchemeOnRedirect,
	}
}

func (h *handlers) NamespaceIndex(c echo.Context) error {
	namespaces, err := h.svc.ListNamespaces(c.Request().Context())
	if err != nil {
		return err
	}

	type data struct {
		Title      string
		Namespaces []string
	}

	return c.Render(http.StatusOK, "namespace-list.html", &data{
		Title:      "Namespace index",
		Namespaces: namespaces,
	})
}

func (h *handlers) ContainerIndex(c echo.Context) error {
	namespace := c.Param("namespace")

	pageParam := c.QueryParam("page")
	var page uint64 = 1

	var err error
	if pageParam != "" {
		page, err = strconv.ParseUint(pageParam, 10, 64)
		if err != nil {
			log.Warnf("malformed page parameter: `%s`", pageParam)
			page = 1
		}
	}

	pagesCount, containers, err := h.svc.ListContainersByPage(c.Request().Context(), namespace, page)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.Render(http.StatusNotFound, notFoundTemplateFilename, nil)
		}
		return err
	}

	type data struct {
		Title       string
		CurrentPage uint64
		PagesCount  uint64
		Namespace   string
		Containers  []models.Container
	}

	return c.Render(http.StatusOK, "container-list.html", &data{
		Title:       fmt.Sprintf("Container index (%s)", namespace),
		CurrentPage: page,
		PagesCount:  pagesCount,
		Namespace:   namespace,
		Containers:  containers,
	})
}

func (h *handlers) VersionIndex(c echo.Context) error {
	namespace := c.Param("namespace")
	container := c.Param("container")

	pageParam := c.QueryParam("page")
	var page uint64 = 1

	var err error
	if pageParam != "" {
		page, err = strconv.ParseUint(pageParam, 10, 64)
		if err != nil {
			log.Warnf("malformed page parameter: `%s`", pageParam)
			page = 1
		}
	}

	pagesCount, versions, err := h.svc.ListPublishedVersionsByPage(c.Request().Context(), namespace, container, page)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.Render(http.StatusNotFound, notFoundTemplateFilename, nil)
		}
		return err
	}

	type data struct {
		Title       string
		CurrentPage uint64
		PagesCount  uint64
		Namespace   string
		Container   string
		Versions    []models.Version
	}

	return c.Render(http.StatusOK, "version-list.html", &data{
		Title:       fmt.Sprintf("Version index (%s/%s)", namespace, container),
		CurrentPage: page,
		PagesCount:  pagesCount,
		Namespace:   namespace,
		Container:   container,
		Versions:    versions,
	})
}

func (h *handlers) ObjectIndex(c echo.Context) error {
	namespace := c.Param("namespace")
	container := c.Param("container")
	version := c.Param("version")

	pageParam := c.QueryParam("page")
	var page uint64 = 1

	var err error
	if pageParam != "" {
		page, err = strconv.ParseUint(pageParam, 10, 64)
		if err != nil {
			log.Warnf("malformed page parameter: `%s`", pageParam)
			page = 1
		}
	}

	pagesCount, objects, err := h.svc.ListObjectsByPage(c.Request().Context(), namespace, container, version, page)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.Render(http.StatusNotFound, notFoundTemplateFilename, nil)
		}
		return err
	}

	type data struct {
		Title       string
		CurrentPage uint64
		PagesCount  uint64
		Namespace   string
		Container   string
		Version     string
		Objects     []string
	}
	return c.Render(http.StatusOK, "object-list.html", &data{
		Title:       fmt.Sprintf("Object index (%s/%s/%s)", namespace, container, version),
		CurrentPage: page,
		PagesCount:  pagesCount,
		Namespace:   namespace,
		Container:   container,
		Version:     version,
		Objects:     objects,
	})
}

func (h *handlers) GetObject(c echo.Context) error {
	namespace := c.Param("namespace")
	container := c.Param("container")
	version := c.Param("version")

	objectParam := c.Param("object")
	key, err := url.PathUnescape(objectParam)
	if err != nil {
		return err
	}

	link, err := h.svc.GetObjectURL(c.Request().Context(), namespace, container, version, key)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return c.Render(http.StatusNotFound, notFoundTemplateFilename, nil)
		}
		return err
	}

	xForwardedScheme := c.Request().Header.Get("X-Forwarded-Scheme")
	xScheme := c.Request().Header.Get("X-Scheme")

	allowedValues := map[string]struct{}{
		"http":  {},
		"https": {},
	}

	_, xForwardedSchemeOk := allowedValues[xForwardedScheme]
	_, xSchemeOk := allowedValues[xScheme]

	if h.preserveSchemeOnRedirect && (xForwardedSchemeOk || xSchemeOk) {
		scheme := xForwardedScheme
		if !xForwardedSchemeOk {
			scheme = xScheme
		}

		u, err := url.Parse(link)
		if err != nil {
			return c.Blob(http.StatusInternalServerError, "text/plain", []byte("error parsing url"))
		}

		u.Scheme = scheme

		link = u.String()
	}

	return c.Redirect(http.StatusFound, link)
}

func (h *handlers) ErrorHandler(err error, c echo.Context) {
	code := 500
	templateFilename := serverErrorTemplateFilename

	v, ok := err.(*echo.HTTPError)
	if ok {
		code = v.Code

		if v.Code == http.StatusNotFound {
			code = http.StatusNotFound
			templateFilename = notFoundTemplateFilename
		}
	}

	type data struct {
		Code    int
		Message string
	}

	if err := c.Render(code, templateFilename, &data{
		Code:    code,
		Message: http.StatusText(code),
	}); err != nil {
		c.Logger().Error(err)
	}
}

func (h *handlers) Register(e *echo.Echo) {
	e.Renderer = &renderer{
		templates: template.Must(
			template.New("base").Funcs(sprig.FuncMap()).ParseGlob(path.Join(h.templateDir, "*.html")),
		),
	}

	e.HTTPErrorHandler = h.ErrorHandler

	e.GET("/", h.NamespaceIndex)
	e.GET("/:namespace/", h.ContainerIndex)
	e.GET("/:namespace/:container/", h.VersionIndex)
	e.GET("/:namespace/:container/:version/", h.ObjectIndex)
	e.GET("/:namespace/:container/:version/:object", h.GetObject)

	e.Static(h.staticDir, "static")
}
