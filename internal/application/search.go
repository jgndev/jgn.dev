package application

import (
	"github.com/jgndev/jgn.dev/internal/contentmanager"
	"github.com/jgndev/jgn.dev/internal/views/components"
	"github.com/labstack/echo/v4"
)

func (a *Application) Search(c echo.Context) error {
	query := c.QueryParam("q")

	if query == "" {
		// Return empty results
		return components.SearchResults([]contentmanager.Post{}).Render(c.Request().Context(), c.Response().Writer)
	}

	// Search posts
	results := a.ContentManager.Search(query)

	// Limit results to 10 for dropdown
	if len(results) > 10 {
		results = results[:10]
	}

	return components.SearchResults(results).Render(c.Request().Context(), c.Response().Writer)
}
