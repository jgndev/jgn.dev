package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// PostsList handles the /posts route and renders a list of all posts, sorted by date in descending order.
func (app *Application) PostsList(c echo.Context) error {
	// Get all posts (already sorted by date, newest first)
	posts := app.ContentManager.GetAll()

	return pages.Posts(posts).Render(c.Request().Context(), c.Response().Writer)
}
