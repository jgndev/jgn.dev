package application

import (
	"net/http"

	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

func (a *Application) CheatsheetDetail(c echo.Context) error {
	slug := c.Param("slug")

	if slug == "" {
		return c.String(http.StatusBadRequest, "Cheatsheet slug is required")
	}

	cheatsheet, exists := a.CheatsheetManager.GetBySlug(slug)
	if !exists {
		return c.String(http.StatusNotFound, "Cheatsheet not found")
	}

	return pages.Cheatsheet(cheatsheet).Render(c.Request().Context(), c.Response().Writer)
}
