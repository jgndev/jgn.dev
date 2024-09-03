package application

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func (a *Application) Health(c echo.Context) error {
	return c.String(http.StatusOK, "healthcheck ok")
}
