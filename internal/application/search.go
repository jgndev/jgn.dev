package application

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/views/lockups"
)

// SearchPosts is the handler for the post searching feature
func (a *Application) SearchPosts(c echo.Context) error {
	q := c.QueryParam("query")
	log.Println(q)

	posts := a.ContentManager.Search(q)

	if len(posts) <= 0 {
		return nil
	}

	return lockups.SearchResults(posts).Render(c.Request().Context(), c.Response().Writer)
}
