package application

import (
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/views/pages"
)

// Plan is the handler for the /plan route
func (a *Application) Plan(c echo.Context) error {
	return pages.Plan().Render(c.Request().Context(), c.Response().Writer)
}
