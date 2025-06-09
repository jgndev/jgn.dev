package application

import (
	"github.com/jgndev/jgn.dev/internal/site"
	"github.com/labstack/echo/v4"
)

// SitemapXML generates an XML sitemap containing all posts and cheatsheets and writes it to the HTTP response as XML.
func (app *Application) SitemapXML(c echo.Context) error {
	posts := app.ContentManager.GetAll()
	cheatsheets := app.CheatsheetManager.GetAll()

	type urlEntry struct {
		Loc     string
		LastMod string
	}

	var urls []urlEntry

	// Add posts
	for _, post := range posts {
		urls = append(urls, urlEntry{
			Loc:     site.URL + "/posts/" + post.Slug,
			LastMod: post.Date.Format("2006-01-02"),
		})
	}
	// Add cheatsheets
	for _, cs := range cheatsheets {
		urls = append(urls, urlEntry{
			Loc:     site.URL + "/cheatsheets/" + cs.Slug,
			LastMod: cs.Date.Format("2006-01-02"),
		})
	}

	xml := `<?xml version="1.0" encoding="UTF-8"?>
	<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	`
	for _, u := range urls {
		xml += "  <url>\n"
		xml += "    <loc>" + u.Loc + "</loc>\n"
		xml += "    <lastmod>" + u.LastMod + "</lastmod>\n"
		xml += "  </url>\n"
	}
	xml += "</urlset>"

	// Set headers before writing
	c.Response().Header().Set("Content-Type", "application/xml; charset=UTF-8")
	c.Response().Header().Set("Cache-Control", "public, max-age=3600")

	// Use Blob instead of String
	return c.Blob(200, "application/xml", []byte(xml))
}
