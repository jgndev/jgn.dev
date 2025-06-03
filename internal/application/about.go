package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// About handles the /about route and renders the About page using the provided context and response writer.
func (app *Application) About(c echo.Context) error {
	return pages.About().Render(c.Request().Context(), c.Response().Writer)
}
