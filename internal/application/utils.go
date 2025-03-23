package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// Utils is the handler for the /utils route
func (a *Application) Utils(c echo.Context) error {
	return pages.Utils().Render(c.Request().Context(), c.Response().Writer)
}
