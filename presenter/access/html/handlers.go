package html

import (
	"html/template"
	"net/http"
	"path"

	echo "github.com/labstack/echo/v4"

	"github.com/teran/archived/service"
)

type Handlers interface {
	ContainerIndex(c echo.Context) error
	VersionIndex(c echo.Context) error

	Register(e *echo.Echo)
}

type handlers struct {
	svc         service.AccessService
	templateDir string
}

func New(svc service.AccessService, templateDir string) Handlers {
	return &handlers{
		svc:         svc,
		templateDir: templateDir,
	}
}

func (h *handlers) ContainerIndex(c echo.Context) error {
	containers, err := h.svc.ListContainers(c.Request().Context())
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "container-list.html", containers)
}

func (h *handlers) VersionIndex(c echo.Context) error {
	container := c.Param("container")
	versions, err := h.svc.ListPublishedVersions(c.Request().Context(), container)
	if err != nil {
		if err == service.ErrNotFound {
			return c.Render(http.StatusNotFound, "404.html", nil)
		}
		return err
	}

	type data struct {
		Container string
		Versions  []string
	}

	return c.Render(http.StatusOK, "version-list.html", &data{
		Container: container,
		Versions:  versions,
	})
}

func (h *handlers) ObjectIndex(c echo.Context) error {
	container := c.Param("container")
	version := c.Param("version")

	objects, err := h.svc.ListObjects(c.Request().Context(), container, version)
	if err != nil {
		if err == service.ErrNotFound {
			return c.Render(http.StatusNotFound, "404.html", nil)
		}
		return err
	}

	type data struct {
		Container string
		Version   string
		Objects   []string
	}
	return c.Render(http.StatusOK, "object-list.html", &data{
		Container: container,
		Version:   version,
		Objects:   objects,
	})
}

func (h *handlers) GetObject(c echo.Context) error {
	container := c.Param("container")
	version := c.Param("version")
	object := c.Param("object")

	url, err := h.svc.GetObjectURL(c.Request().Context(), container, version, object)
	if err != nil {
		if err == service.ErrNotFound {
			return c.Render(http.StatusNotFound, "404.html", nil)
		}
		return err
	}

	return c.Redirect(http.StatusFound, url)
}

func (h *handlers) ErrorHandler(err error, c echo.Context) {
	code := 500
	templateFilename := "5xx.html"

	v, ok := err.(*echo.HTTPError)
	if ok {
		code = v.Code

		switch v.Code {
		case http.StatusNotFound:
			code = http.StatusNotFound
			templateFilename = "404.html"
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
		templates: template.Must(template.ParseGlob(path.Join(h.templateDir, "*.html"))),
	}

	e.HTTPErrorHandler = h.ErrorHandler

	e.GET("/", h.ContainerIndex)
	e.GET("/:container/", h.VersionIndex)
	e.GET("/:container/:version/", h.ObjectIndex)
	e.GET("/:container/:version/:object", h.GetObject)
}
