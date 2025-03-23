package application

import (
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/views/pages"
)

func (a *Application) Home(c echo.Context) error {
	posts := a.ContentManager.GetRecent(3)
	return pages.Home(posts).Render(c.Request().Context(), c.Response().Writer)
}
