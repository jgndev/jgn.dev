package application

import (
	"log"

	"github.com/jgndev/jgn.dev/internal/contentmanager"
	"github.com/jgndev/jgn.dev/internal/site"
	"github.com/labstack/echo/v4"
)

type Application struct {
	ContentManager    *contentmanager.ContentManager
	CheatsheetManager *contentmanager.CheatsheetManager
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

// About renders the About page
func (app *Application) About(c echo.Context) error {
	return AboutHandler(c)
}
