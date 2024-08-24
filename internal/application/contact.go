package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// Contact is the handler for the /contact route
func (a *Application) Contact(c echo.Context) error {
	return pages.Contact().Render(c.Request().Context(), c.Response().Writer)
}
