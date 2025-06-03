package application

import (
	"github.com/jgndev/jgn.dev/internal/contentmanager"
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// SearchPage handles the search functionality for blog posts. It processes the query parameter "q" and renders search results.
func (app *Application) SearchPage(c echo.Context) error {
	query := c.QueryParam("q")

	var results []contentmanager.Post

	if query != "" {
		results = app.ContentManager.Search(query)
	}

	return pages.SearchPage(query, results).Render(c.Request().Context(), c.Response().Writer)
}

// CheatsheetSearchPage handles the search functionality for cheatsheets, processing the "q" query parameter and rendering results.
func (app *Application) CheatsheetSearchPage(c echo.Context) error {
	query := c.QueryParam("q")

	var results []contentmanager.Cheatsheet

	if query != "" {
		results = app.CheatsheetManager.Search(query)
	}

	return pages.CheatsheetSearchPage(query, results).Render(c.Request().Context(), c.Response().Writer)
}
