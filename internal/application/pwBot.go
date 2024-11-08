package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// PwBot is the handler for the /pwbot route
func (a *Application) PwBot(c echo.Context) error {
	return pages.PwBot().Render(c.Request().Context(), c.Response().Writer)
}
