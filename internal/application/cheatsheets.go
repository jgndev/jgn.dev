package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// CheatsheetsList handles the /cheatsheets route and renders a list of all cheatsheets, sorted by date (newest first).
func (app *Application) CheatsheetsList(c echo.Context) error {
	// Get all cheatsheets (already sorted by date, newest first)
	cheatsheets := app.CheatsheetManager.GetAll()

	return pages.Cheatsheets(cheatsheets).Render(c.Request().Context(), c.Response().Writer)
}
