package application

import (
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/views/pages"
)

func (a *Application) Home(c echo.Context) error {
	// Get recent posts (latest 6)
	recentPosts := a.ContentManager.GetRecent(6)
	
	return pages.Home(recentPosts).Render(c.Request().Context(), c.Response().Writer)
}
