package application

import (
	"net/http"

	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// CheatsheetDetail handles the /cheatsheets/:slug route.
// It fetches a cheatsheet by slug and renders the detail page, or returns an error if not found.
func (app *Application) CheatsheetDetail(c echo.Context) error {
	slug := c.Param("slug")

	if slug == "" {
		return c.String(http.StatusBadRequest, "Cheatsheet slug is required")
	}

	cheatsheet, exists := app.CheatsheetManager.GetBySlug(slug)
	if !exists {
		return c.String(http.StatusNotFound, "Cheatsheet not found")
	}

	return pages.Cheatsheet(cheatsheet).Render(c.Request().Context(), c.Response().Writer)
}
