package application

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"time"

	_ "time/tzdata"
)

func (a *Application) GetTime(c echo.Context) error {

	log.Println("GetTime function called")

	zones, count := time.Now().In(time.Local).Zone()
	log.Printf("Available time zones: %v, %d total", zones, count)

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
