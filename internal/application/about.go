package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// About is the handler for the /about route
func (a *Application) About(c echo.Context) error {
	return pages.About().Render(c.Request().Context(), c.Response().Writer)
}
