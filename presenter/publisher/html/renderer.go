package html

import (
	"html/template"
	"io"

	echo "github.com/labstack/echo/v4"
)

type renderer struct {
	templates *template.Template
}

func (r *renderer) Render(w io.Writer, name string, data any, c echo.Context) error {
	return r.templates.ExecuteTemplate(w, name, data)
}
