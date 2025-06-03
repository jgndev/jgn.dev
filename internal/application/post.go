package application

import (
	"net/http"

	"github.com/jgndev/jgn.dev/internal/views/pages"
	"github.com/labstack/echo/v4"
)

// PostDetail handles the /posts/:slug route and renders the blog post detail page or returns an error if not found.
func (app *Application) PostDetail(c echo.Context) error {
	slug := c.Param("slug")

	if slug == "" {
		return c.String(http.StatusBadRequest, "Post slug is required")
	}

	post, exists := app.ContentManager.GetBySlug(slug)
	if !exists {
		return c.String(http.StatusNotFound, "Post not found")
	}

	return pages.Post(post).Render(c.Request().Context(), c.Response().Writer)
}
