package application

import (
	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

func AboutHandler(c echo.Context) error {
	return pages.About().Render(c.Request().Context(), c.Response().Writer)
}
