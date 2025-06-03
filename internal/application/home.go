package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// Home handles the root route and renders the home page with the latest 6 posts.
func (app *Application) Home(c echo.Context) error {
	// Get recent posts (latest 6)
	recentPosts := app.ContentManager.GetRecent(6)

	return pages.Home(recentPosts).Render(c.Request().Context(), c.Response().Writer)
}
