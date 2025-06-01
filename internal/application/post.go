package application

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/views/pages"
)

func (a *Application) PostDetail(c echo.Context) error {
	slug := c.Param("slug")
	
	if slug == "" {
		return c.String(http.StatusBadRequest, "Post slug is required")
	}
	
	post, exists := a.ContentManager.GetBySlug(slug)
	if !exists {
		return c.String(http.StatusNotFound, "Post not found")
	}
	
	return pages.Post(post).Render(c.Request().Context(), c.Response().Writer)
} 