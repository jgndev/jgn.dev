package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// Plan is the handler for the /plan route
func (a *Application) Plan(c echo.Context) error {
	return pages.Plan().Render(c.Request().Context(), c.Response().Writer)
}
