package application

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jgndev/jgn.dev/internal/site"
	"github.com/labstack/echo/v4"
)

// GitHubWebhookPayload represents the data structure of a payload received from a GitHub webhook event.
type GitHubWebhookPayload struct {
	Ref     string `json:"ref"`
	Commits []struct {
		Added    []string `json:"added"`
		Modified []string `json:"modified"`
		Removed  []string `json:"removed"`
	} `json:"commits"`
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// WebhookHandler handles incoming GitHub webhook requests, verifying the signature and processing relevant events.
func (app *Application) WebhookHandler(c echo.Context) error {
	// Get the webhook secret from the environment
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		log.Printf("Webhook received but no GITHUB_WEBHOOK_SECRET configured")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "webhook secret not configured",
		})
	}

	// Read the request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Failed to read webhook body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "failed to read request body",
		})
	}

	// Verify the webhook signature
	signature := c.Request().Header.Get("X-Hub-Signature-256")
	if !verifyWebhookSignature(body, signature, secret) {
		log.Printf("Invalid webhook signature")
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid signature",
		})
	}

	// Parse the webhook payload
	var payload GitHubWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Failed to parse webhook payload: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "failed to parse payload",
		})
	}

	// Check if this is a push to the main branch
	if payload.Ref != "refs/heads/main" && payload.Ref != "refs/heads/master" {
		log.Printf("Webhook received for non-main branch: %s", payload.Ref)
		return c.JSON(http.StatusOK, map[string]string{
			"message": "ignoring non-main branch push",
		})
	}

	// Check if any Markdown files were added or modified
	hasMarkdownChanges := false
	for _, commit := range payload.Commits {
		for _, file := range append(commit.Added, commit.Modified...) {
			if strings.HasSuffix(strings.ToLower(file), ".md") {
				hasMarkdownChanges = true
				log.Printf("Detected markdown file change: %s", file)
				break
			}
		}
		if hasMarkdownChanges {
			break
		}
	}

	if !hasMarkdownChanges {
		log.Printf("Webhook received but no markdown files changed")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "no markdown files changed",
		})
	}

	// Determine which content manager to refresh based on repository
	repoName := payload.Repository.Name
	log.Printf("Refreshing content due to webhook from %s (repo: %s)", payload.Repository.FullName, repoName)
	
	// Import site package to access repository names
	var refreshErr error
	refreshed := false
	
	// Check if this is the posts repository
	if repoName == site.PostRepoName {
		log.Printf("Detected posts repository, refreshing ContentManager")
		refreshErr = app.ContentManager.RefreshContent()
		refreshed = true
	}
	
	// Check if this is the cheatsheets repository
	if repoName == site.CheatsheetRepoName {
		log.Printf("Detected cheatsheets repository, refreshing CheatsheetManager")
		refreshErr = app.CheatsheetManager.RefreshContent()
		refreshed = true
	}
	
	// If repository wasn't recognized, refresh both managers as fallback
	if !refreshed {
		log.Printf("WARNING: Unknown repository '%s', refreshing both managers as fallback", repoName)
		
		// Try to refresh posts first
		if err := app.ContentManager.RefreshContent(); err != nil {
			log.Printf("Failed to refresh posts during fallback: %v", err)
			refreshErr = err
		} else {
			log.Printf("Successfully refreshed posts during fallback")
		}
		
		// Try to refresh cheatsheets
		if err := app.CheatsheetManager.RefreshContent(); err != nil {
			log.Printf("Failed to refresh cheatsheets during fallback: %v", err)
			if refreshErr == nil {
				refreshErr = err
			}
		} else {
			log.Printf("Successfully refreshed cheatsheets during fallback")
		}
	}
	
	// Handle refresh errors
	if refreshErr != nil {
		log.Printf("Failed to refresh content for repository %s: %v", repoName, refreshErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to refresh content",
			"repository": repoName,
		})
	}

	log.Printf("Successfully refreshed content from webhook for repository: %s", repoName)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "content refreshed successfully",
		"repository": repoName,
	})
}

// verifyWebhookSignature validates a webhook payload signature against the expected HMAC-SHA256 signature.
func verifyWebhookSignature(body []byte, signature, secret string) bool {
	if signature == "" {
		return false
	}

	// Remove the "sha256=" prefix
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	signature = signature[7:]

	// Calculate the expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
