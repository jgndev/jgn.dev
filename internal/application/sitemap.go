package application

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/models"
	"net/http"
	"time"
)

func (a *Application) SiteMap(c echo.Context) error {
	sd := []models.SitemapData{
		{
			Loc:        "https://stratocraft.dev",
			LastMod:    sitemapDate(),
			ChangeFreq: "weekly",
			Priority:   "1.0",
		},
		{
			Loc:        "https://stratocraft.dev/posts",
			LastMod:    sitemapDate(),
			ChangeFreq: "weekly",
			Priority:   "1.0",
		},
		{
			Loc:        "https://stratocraft.dev/about",
			LastMod:    sitemapDate(),
			ChangeFreq: "weekly",
			Priority:   "1.0",
		},
		{
			Loc:        "https://stratocraft.dev/contact",
			LastMod:    sitemapDate(),
			ChangeFreq: "weekly",
			Priority:   "1.0",
		},
	}

	header := `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`
	footer := `</urlset>`

	var xml string
	xml = xml + header

	for _, s := range sd {
		xml += fmt.Sprintf(`
		<url>
		<loc>%s</loc>
		<lastmod>%s</lastmod>
		<changefreq>%s</changefreq>
		<priority>%s</priority>
		</url>`, s.Loc, s.LastMod, s.ChangeFreq, s.Priority)
	}

	xml = xml + footer

	return c.XML(http.StatusOK, xml)
}

func sitemapDate() string {
	return time.Now().UTC().Format(time.RFC3339)
}
