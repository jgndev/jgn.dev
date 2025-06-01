package application

import (
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/views/pages"
)

func AboutHandler(c echo.Context) error {
	return pages.About().Render(c.Request().Context(), c.Response().Writer)
} 