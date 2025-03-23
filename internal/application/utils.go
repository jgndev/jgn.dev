package application

import (
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/views/pages"
)

// Utils is the handler for the /utils route
func (a *Application) Utils(c echo.Context) error {
	return pages.Utils().Render(c.Request().Context(), c.Response().Writer)
}
