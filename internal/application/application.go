package application

import (
	"log"

	"github.com/jgndev/jgn.dev/internal/contentmanager"
	"github.com/jgndev/jgn.dev/internal/site"
	"github.com/labstack/echo/v4"
)

// Application represents the core structure of the application, managing blog post and cheatsheet content through dedicated managers.
type Application struct {
	ContentManager    *contentmanager.ContentManager    // Manages blog post content
	CheatsheetManager *contentmanager.CheatsheetManager // Manages cheatsheet content
}

// New initializes and returns a pointer to an Application instance, setting up content and cheatsheet managers.
func New() *Application {
	repoOwner := site.PostRepoOwner
	if len(repoOwner) <= 0 {
		log.Fatal("PostRepoOwner must be set in site.go to the account name that owns the repo on github.com")
	}

	repoName := site.PostRepoName
	if len(repoName) <= 0 {
		log.Fatal("PostRepoName must be set in site.go to the repo name that has the posts on github.com")
	}

	cm := contentmanager.NewContentManager(repoOwner, repoName)
	if err := cm.RefreshContent(); err != nil {
		log.Printf("Failed to load initial content: %v", err)
	}

	// Initialize cheatsheet manager
	cheatsheetOwner := site.CheatsheetRepoOwner
	if len(cheatsheetOwner) <= 0 {
		log.Fatal("CheatsheetRepoOwner must be set in site.go to the account name that owns the cheatsheets repo on github.com")
	}

	cheatsheetName := site.CheatsheetRepoName
	if len(cheatsheetName) <= 0 {
		log.Fatal("CheatsheetRepoName must be set in site.go to the repo name that has the cheatsheets on github.com")
	}

	csm := contentmanager.NewCheatsheetManager(cheatsheetOwner, cheatsheetName)
	if err := csm.RefreshContent(); err != nil {
		log.Printf("Failed to load initial cheatsheets: %v", err)
	}

	return &Application{
		ContentManager:    cm,
		CheatsheetManager: csm,
	}
}

// SitemapXML generates an XML sitemap containing all posts and cheatsheets, and writes it to the HTTP response as XML.
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

	xml := `<?xml version="1.0" encoding="UTF-8"?>\n` +
		`<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n`
	for _, u := range urls {
		xml += "  <url>\n"
		xml += "    <loc>" + u.Loc + "</loc>\n"
		xml += "    <lastmod>" + u.LastMod + "</lastmod>\n"
		xml += "  </url>\n"
	}
	xml += "</urlset>\n"

	c.Response().Header().Set("Content-Type", "application/xml")
	return c.String(200, xml)
}
