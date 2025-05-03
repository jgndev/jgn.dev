package application

import (
	"bytes"
	"encoding/xml"
	"github.com/jgndev/jgn.dev/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

func (a *Application) SiteMap(c echo.Context) error {
	// Define the header and footer
	header := `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`
	footer := `</urlset>`

	// Define the sitemap data
	urls := []models.SitemapURL{
		{
			Loc:        "https://jgn.dev",
			LastMod:    sitemapDate(),
			ChangeFreq: "weekly",
			Priority:   "1.0",
		},
		{
			Loc:        "https://jgn.dev/plan",
			LastMod:    sitemapDate(),
			ChangeFreq: "weekly",
			Priority:   "1.0",
		},
		{
			Loc:        "https://jgn.dev/utils",
			LastMod:    sitemapDate(),
			ChangeFreq: "weekly",
			Priority:   "1.0",
		},
		{
			Loc:        "https://jgn.dev/posts",
			LastMod:    sitemapDate(),
			ChangeFreq: "weekly",
			Priority:   "1.0",
		},
		{
			Loc:        "https://jgn.dev/about",
			LastMod:    sitemapDate(),
			ChangeFreq: "weekly",
			Priority:   "1.0",
		},
		{
			Loc:        "https://jgn.dev/contact",
			LastMod:    sitemapDate(),
			ChangeFreq: "weekly",
			Priority:   "1.0",
		},
	}

	// Use a buffer to build the XML
	var buf bytes.Buffer
	buf.WriteString(xml.Header) // Add <?xml version="1.0" encoding="UTF-8"?>
	buf.WriteString(header)
	buf.WriteString("\n") // Optional: for readability

	// Generate XML for each <url> entry with indentation
	for _, u := range urls {
		urlData, err := xml.MarshalIndent(u, "  ", "  ") // Indent with 2 spaces
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to generate sitemap")
		}
		buf.WriteString(string(urlData))
		buf.WriteString("\n") // Optional: for readability
	}

	buf.WriteString(footer)

	// Set content type and return the XML
	c.Response().Header().Set("Content-Type", "application/xml")
	return c.Blob(http.StatusOK, "application/xml", buf.Bytes())
}

func sitemapDate() string {
	return time.Now().UTC().Format("2006-01-02") // Use YYYY-MM-DD for sitemap
}
