package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// Home is the handler for the / route
func (a *Application) Home(c echo.Context) error {
	posts := a.Blog.GetRecent(3)
	return pages.Home(posts).Render(c.Request().Context(), c.Response().Writer)
}
