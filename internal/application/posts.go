package application

import (
	"net/http"

	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// Posts is the handler for the /posts route
func (a *Application) Posts(c echo.Context) error {
	posts := a.Blog.GetAll()
	return pages.Posts(posts).Render(c.Request().Context(), c.Response().Writer)
}

// Post is the handler for an individual post
func (a *Application) Post(c echo.Context) error {
	slug := c.Param("slug")

	post, ok := a.Blog.GetBySlug(slug)
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return pages.Post(*post).Render(c.Request().Context(), c.Response().Writer)
}
