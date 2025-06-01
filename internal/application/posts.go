package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

func (a *Application) PostsList(c echo.Context) error {
	// Get all posts (already sorted by date, newest first)
	posts := a.ContentManager.GetAll()

	return pages.Posts(posts).Render(c.Request().Context(), c.Response().Writer)
}
