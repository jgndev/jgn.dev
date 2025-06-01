package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

func (a *Application) Home(c echo.Context) error {
	// Get recent posts (latest 6)
	recentPosts := a.ContentManager.GetRecent(6)

	return pages.Home(recentPosts).Render(c.Request().Context(), c.Response().Writer)
}
