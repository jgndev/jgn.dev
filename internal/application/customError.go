package application

import (
	"errors"
	"net/http"

	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// CustomErrorHandler is the handler for custom errors
func (a *Application) CustomErrorHandler(err error, c echo.Context) {
	var httpError *echo.HTTPError
	if errors.As(err, &httpError) {
		status := httpError.Code
		switch status {
		case http.StatusNotFound:
			_ = pages.NotFound().Render(c.Request().Context(), c.Response().Writer)
		case http.StatusInternalServerError:
			_ = pages.ServerError().Render(c.Request().Context(), c.Response().Writer)
		default:
			_ = c.HTML(status, "oops")
		}
	}
}
