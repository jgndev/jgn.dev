package application

import (
	"github.com/jgndev/jgn.dev/internal/contentmanager"
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

func (a *Application) SearchPage(c echo.Context) error {
	query := c.QueryParam("q")

	var results []contentmanager.Post

	if query != "" {
		results = a.ContentManager.Search(query)
	}

	return pages.SearchPage(query, results).Render(c.Request().Context(), c.Response().Writer)
}
