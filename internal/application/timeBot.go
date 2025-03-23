package application

import (
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/views/pages"
)

// TimeBot is the handler for the /timebot route
func (a *Application) TimeBot(c echo.Context) error {
	return pages.TimeBot().Render(c.Request().Context(), c.Response().Writer)
}
