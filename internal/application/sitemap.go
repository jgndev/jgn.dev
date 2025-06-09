// package application

// import (
// 	"github.com/jgndev/jgn.dev/internal/site"
// 	"github.com/labstack/echo/v4"
// )

// // SitemapXML generates an XML sitemap containing all posts and cheatsheets and writes it to the HTTP response as XML.
// func (app *Application) SitemapXML(c echo.Context) error {
// 	posts := app.ContentManager.GetAll()
// 	cheatsheets := app.CheatsheetManager.GetAll()

// 	type urlEntry struct {
// 		Loc     string
// 		LastMod string
// 	}

// 	var urls []urlEntry

// 	// Add posts
// 	for _, post := range posts {
// 		urls = append(urls, urlEntry{
// 			Loc:     site.URL + "/posts/" + post.Slug,
// 			LastMod: post.Date.Format("2006-01-02"),
// 		})
// 	}
// 	// Add cheatsheets
// 	for _, cs := range cheatsheets {
// 		urls = append(urls, urlEntry{
// 			Loc:     site.URL + "/cheatsheets/" + cs.Slug,
// 			LastMod: cs.Date.Format("2006-01-02"),
// 		})
// 	}

// 	xml := `<?xml version="1.0" encoding="UTF-8"?>
// 	<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
// 	`
// 	for _, u := range urls {
// 		xml += "  <url>\n"
// 		xml += "    <loc>" + u.Loc + "</loc>\n"
// 		xml += "    <lastmod>" + u.LastMod + "</lastmod>\n"
// 		xml += "  </url>\n"
// 	}
// 	xml += "</urlset>"

// 	// Set headers before writing
// 	c.Response().Header().Set("Content-Type", "application/xml; charset=UTF-8")
// 	c.Response().Header().Set("Cache-Control", "public, max-age=3600")

// 	// Use Blob instead of String
// 	return c.Blob(200, "application/xml", []byte(xml))
// }

package application

import (
	"encoding/xml"
	"time"

	"github.com/jgndev/jgn.dev/internal/site"
	"github.com/labstack/echo/v4"
)

// URLEntry represents a single URL in the sitemap
type URLEntry struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	LastMod    string   `xml:"lastmod"`
	ChangeFreq string   `xml:"changefreq,omitempty"`
	Priority   string   `xml:"priority,omitempty"`
}

// URLSet represents the root element of the sitemap
type URLSet struct {
	XMLName xml.Name   `xml:"urlset"`
	Xmlns   string     `xml:"xmlns,attr"`
	URLs    []URLEntry `xml:"url"`
}

// SitemapXML generates an XML sitemap containing all posts and cheatsheets
func (app *Application) SitemapXML(c echo.Context) error {
	posts := app.ContentManager.GetAll()
	cheatsheets := app.CheatsheetManager.GetAll()

	// Create URLSet with proper namespace
	urlSet := URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]URLEntry, 0),
	}

	// Add homepage
	urlSet.URLs = append(urlSet.URLs, URLEntry{
		Loc:        site.URL,
		LastMod:    time.Now().Format("2006-01-02"),
		ChangeFreq: "weekly",
		Priority:   "1.0",
	})

	// Add main sections
	urlSet.URLs = append(urlSet.URLs, URLEntry{
		Loc:        site.URL + "/posts",
		LastMod:    time.Now().Format("2006-01-02"),
		ChangeFreq: "daily",
		Priority:   "0.9",
	})

	urlSet.URLs = append(urlSet.URLs, URLEntry{
		Loc:        site.URL + "/cheatsheets",
		LastMod:    time.Now().Format("2006-01-02"),
		ChangeFreq: "weekly",
		Priority:   "0.9",
	})

	urlSet.URLs = append(urlSet.URLs, URLEntry{
		Loc:        site.URL + "/about",
		LastMod:    time.Now().Format("2006-01-02"),
		ChangeFreq: "monthly",
		Priority:   "0.8",
	})

	// Add posts
	for _, post := range posts {
		urlSet.URLs = append(urlSet.URLs, URLEntry{
			Loc:        site.URL + "/posts/" + post.Slug,
			LastMod:    post.Date.Format("2006-01-02"),
			ChangeFreq: "never",
			Priority:   "0.7",
		})
	}

	// Add cheatsheets
	for _, cs := range cheatsheets {
		urlSet.URLs = append(urlSet.URLs, URLEntry{
			Loc:        site.URL + "/cheatsheets/" + cs.Slug,
			LastMod:    cs.Date.Format("2006-01-02"),
			ChangeFreq: "monthly",
			Priority:   "0.8",
		})
	}

	// Marshal to XML with proper header
	xmlData, err := xml.MarshalIndent(urlSet, "", "  ")
	if err != nil {
		return echo.NewHTTPError(500, "Failed to generate sitemap")
	}

	// Add XML declaration
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(xmlData)

	// Set proper headers
	c.Response().Header().Set("Content-Type", "application/xml; charset=UTF-8")
	c.Response().Header().Set("Cache-Control", "public, max-age=3600")

	// IMPORTANT: Use String() instead of Blob to avoid gzip issues
	return c.String(200, xmlContent)
}
