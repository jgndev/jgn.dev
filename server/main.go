// package main

// import (
// 	"net/http"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/jgndev/jgn.dev/internal/application"
// 	"github.com/labstack/echo/v4"
// 	"github.com/labstack/echo/v4/middleware"
// )

// // cacheMiddleware adds appropriate cache headers for static assets
// func cacheMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		path := c.Request().URL.Path

// 		// Only apply caching to static assets
// 		if !strings.HasPrefix(path, "/public/") &&
// 			path != "/favicon.ico" &&
// 			path != "/robots.txt" {
// 			return next(c)
// 		}

// 		// Determine cache duration based on file type
// 		var maxAge time.Duration

// 		switch {
// 		case strings.HasSuffix(path, ".woff2") || strings.HasSuffix(path, ".woff") ||
// 			strings.HasSuffix(path, ".ttf") || strings.HasSuffix(path, ".otf"):
// 			// Fonts: 1 year (they rarely change)
// 			maxAge = 365 * 24 * time.Hour

// 		case strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js"):
// 			// CSS and JS: 30 days (may change with updates)
// 			maxAge = 30 * 24 * time.Hour

// 		case strings.HasSuffix(path, ".ico") || strings.HasSuffix(path, ".png") ||
// 			strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") ||
// 			strings.HasSuffix(path, ".gif") || strings.HasSuffix(path, ".svg") ||
// 			strings.HasSuffix(path, ".webp"):
// 			// Images: 30 days
// 			maxAge = 30 * 24 * time.Hour

// 		case strings.HasSuffix(path, ".txt"):
// 			// Text files like robots.txt: 1 day (might need updates)
// 			maxAge = 24 * time.Hour

// 		default:
// 			// Default for other static assets: 7 days
// 			maxAge = 7 * 24 * time.Hour
// 		}

// 		// Set cache headers
// 		if maxAge > 0 {
// 			maxAgeSeconds := int(maxAge.Seconds())
// 			c.Response().Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAgeSeconds))
// 			c.Response().Header().Set("Expires", time.Now().Add(maxAge).UTC().Format(http.TimeFormat))
// 			// Add ETag for better cache validation
// 			c.Response().Header().Set("ETag", `"static-asset"`)
// 		}

// 		return next(c)
// 	}
// }

// func main() {
// 	e := echo.New()

// 	// Configure middleware
// 	e.Use(middleware.Recover())
// 	e.Use(middleware.CORS())
// 	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
// 		Level: 5,
// 	}))

// 	// Apply cache middleware to all requests (it will only set headers for static assets)
// 	e.Use(cacheMiddleware)

// 	// Static assets bundling - back to original setup
// 	e.Static("/public", "public")
// 	e.File("/favicon.ico", "public/img/favicon.ico")
// 	e.File("/robots.txt", "public/txt/robots.txt")

// 	app := application.New()

// 	// Routes
// 	e.GET("/", app.Home)
// 	e.GET("/posts", app.PostsList)
// 	e.GET("/search", app.SearchPage)
// 	e.GET("/posts/:slug", app.PostDetail)
// 	e.GET("/cheatsheets", app.CheatsheetsList)
// 	e.GET("/cheatsheets/search", app.CheatsheetSearchPage)
// 	e.GET("/cheatsheets/:slug", app.CheatsheetDetail)
// 	e.GET("/about", app.About)

// 	// Sitemap
// 	e.GET("/sitemap.xml", app.SitemapXML)

// 	// Webhook for automatic content updates
// 	e.POST("/webhook/github", app.WebhookHandler)

// 	// Start the application
// 	e.Logger.Fatal(e.Start(":8080"))
// }

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jgndev/jgn.dev/internal/application"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// cacheMiddleware adds appropriate cache headers for static assets
func cacheMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().URL.Path

		// Only apply caching to static assets
		if !strings.HasPrefix(path, "/public/") &&
			path != "/favicon.ico" &&
			path != "/robots.txt" &&
			path != "/sitemap.xml" {
			return next(c)
		}

		// Determine cache duration based on file type
		var maxAge time.Duration

		switch {
		case strings.HasSuffix(path, ".woff2") || strings.HasSuffix(path, ".woff") ||
			strings.HasSuffix(path, ".ttf") || strings.HasSuffix(path, ".otf"):
			// Fonts: 1 year (they rarely change)
			maxAge = 365 * 24 * time.Hour

		case strings.HasSuffix(path, ".css") || strings.HasSuffix(path, ".js"):
			// CSS and JS: 30 days (may change with updates)
			maxAge = 30 * 24 * time.Hour

		case strings.HasSuffix(path, ".ico") || strings.HasSuffix(path, ".png") ||
			strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") ||
			strings.HasSuffix(path, ".gif") || strings.HasSuffix(path, ".svg") ||
			strings.HasSuffix(path, ".webp"):
			// Images: 30 days
			maxAge = 30 * 24 * time.Hour

		case strings.HasSuffix(path, ".txt"):
			// Text files like robots.txt: 1 day (might need updates)
			maxAge = 24 * time.Hour

		case strings.HasSuffix(path, ".xml"):
			// XML files like sitemap: 1 hour (for SEO freshness)
			maxAge = 1 * time.Hour

		default:
			// Default for other static assets: 7 days
			maxAge = 7 * 24 * time.Hour
		}

		// Set cache headers
		if maxAge > 0 {
			maxAgeSeconds := int(maxAge.Seconds())
			c.Response().Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAgeSeconds))
			c.Response().Header().Set("Expires", time.Now().Add(maxAge).UTC().Format(http.TimeFormat))
			// Add ETag for better cache validation
			c.Response().Header().Set("ETag", `"static-asset"`)
		}

		return next(c)
	}
}

// validateEnvironment checks critical environment variables and logs warnings
func validateEnvironment() {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Println("WARNING: GITHUB_TOKEN not set - GitHub API requests will be rate limited (60/hour vs 5000/hour)")
		log.Println("         This may cause webhook failures during high traffic periods")
		log.Println("         Set GITHUB_TOKEN environment variable with a GitHub personal access token")
	} else {
		log.Println("✓ GITHUB_TOKEN configured - GitHub API rate limit: 5000/hour")
	}
	
	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Println("WARNING: GITHUB_WEBHOOK_SECRET not set - webhook endpoint will reject all requests")
		log.Println("         Set GITHUB_WEBHOOK_SECRET environment variable to enable webhook functionality")
	} else {
		log.Println("✓ GITHUB_WEBHOOK_SECRET configured - webhook endpoint secured")
	}
}

func main() {
	// Validate critical environment variables
	validateEnvironment()
	
	e := echo.New()

	// Configure middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Configure Gzip with skipper to exclude sitemap.xml
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
		Skipper: func(c echo.Context) bool {
			// Skip gzip for sitemap.xml to avoid XML parsing issues
			return c.Path() == "/sitemap.xml"
		},
	}))

	// Apply cache middleware to all requests (it will only set headers for static assets)
	e.Use(cacheMiddleware)

	// Static assets bundling
	e.Static("/public", "public")
	e.File("/favicon.ico", "public/img/favicon.ico")
	e.File("/robots.txt", "public/txt/robots.txt")

	app := application.New()

	// Routes
	e.GET("/", app.Home)
	e.GET("/posts", app.PostsList)
	e.GET("/search", app.SearchPage)
	e.GET("/posts/:slug", app.PostDetail)
	e.GET("/cheatsheets", app.CheatsheetsList)
	e.GET("/cheatsheets/search", app.CheatsheetSearchPage)
	e.GET("/cheatsheets/:slug", app.CheatsheetDetail)
	e.GET("/about", app.About)

	// Sitemap
	e.GET("/sitemap.xml", app.SitemapXML)

	// Webhook for automatic content updates
	e.POST("/webhook/github", app.WebhookHandler)

	// Start the application
	e.Logger.Fatal(e.Start(":8080"))
}
