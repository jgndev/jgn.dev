package application

import (
	"log"
	"os"

	"github.com/stratocraft/stratocraft.dev/internal/contentmanager"
)

type Application struct {
	ContentManager *contentmanager.ContentManager
}

func NewApplication() *Application {
	repoOwner := os.Getenv("GITHUB_OWNER")
	if repoOwner == "" {
		log.Fatal("GITHUB_OWNER environment variable not set!")
	}

	repoName := os.Getenv("GITHUB_REPO")
	if repoName == "" {
		log.Fatal("GITHUB_REPO_NAME environment variable not set!")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable not set!")
	}

	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("GITHUB_WEBHOOK_SECRET environment variable not set!")
	}

	cm := contentmanager.NewContentManger(repoOwner, repoName, token, secret)
	if err := cm.RefreshContent(); err != nil {
		log.Printf("Failed to load initial content: %v", err)
	}

	return &Application{
		ContentManager: cm,
	}
}
