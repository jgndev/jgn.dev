package application

import (
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/views/pages"
)

// PwBot is the handler for the /pwbot route
func (a *Application) PwBot(c echo.Context) error {
	return pages.PwBot().Render(c.Request().Context(), c.Response().Writer)
}
