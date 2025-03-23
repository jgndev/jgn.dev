package application

import (
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/views/cards"
	"net/http"
	"time"
)

func (a *Application) TimeUpdate(c echo.Context) error {
	return cards.TimeUpdate().Render(c.Request().Context(), c.Response().Writer)
}

func (a *Application) GetTime(c echo.Context) error {

	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		loc = time.UTC
	}
	currentTime := time.Now().In(loc).Format("3:04 PM CST")
	return c.String(http.StatusOK, currentTime)
}
