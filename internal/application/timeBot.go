package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// TimeBot is the handler for the /timebot route
func (a *Application) TimeBot(c echo.Context) error {
	return pages.TimeBot().Render(c.Request().Context(), c.Response().Writer)
}
