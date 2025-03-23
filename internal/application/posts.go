package application

import (
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/views/pages"
	"net/http"
)

func (a *Application) Posts(c echo.Context) error {
	posts := a.ContentManager.GetAll()
	return pages.Posts(posts).Render(c.Request().Context(), c.Response().Writer)
}

func (a *Application) Post(c echo.Context) error {
	slug := c.Param("slug")

	post, ok := a.ContentManager.GetBySlug(slug)
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	//chromaCSS, _ := styles.Get("catppuccin-mocha").Builder().Build()

	//return pages.Post(post, chromaCSS).Render(c.Request().Context(), c.Response().Writer)
	return pages.Post(post).Render(c.Request().Context(), c.Response().Writer)
}
