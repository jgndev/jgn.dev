package application

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (a *Application) Health(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
