package application

import (
	"github.com/labstack/echo/v4"
	"github.com/stratocraft/stratocraft.dev/internal/contentmanager"
	"github.com/stratocraft/stratocraft.dev/internal/site"
	"log"
)

type Application struct {
	ContentManager *contentmanager.ContentManager
}

func New() *Application {
	repoOwner := site.PostRepoOwner
	if len(repoOwner) <= 0 {
		log.Fatal("PostRepoOwner must be set in site.go to the account name that owns the repo on github.com")
	}

	repoName := site.PostRepoName
	if len(repoName) <= 0 {
		log.Fatal("PostRepoName must be set in site.go to the repo name that has the posts on github.com")
	}

	cm := contentmanager.New(repoOwner, repoName)
	if err := cm.RefreshContent(); err != nil {
		log.Printf("Failed to load initial content: %v", err)
	}

	return &Application{
		ContentManager: cm,
	}
}

// About renders the About page
func (app *Application) About(c echo.Context) error {
	return AboutHandler(c)
}
