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

	"github.com/labstack/echo/v4"
)

// GitHubWebhookPayload represents the relevant parts of a GitHub push webhook
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

// WebhookHandler handles GitHub webhook events
func (app *Application) WebhookHandler(c echo.Context) error {
	// Get the webhook secret from environment
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

	// Check if any markdown files were added or modified
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

	// Refresh content from GitHub
	log.Printf("Refreshing content due to webhook from %s", payload.Repository.FullName)
	if err := app.ContentManager.RefreshContent(); err != nil {
		log.Printf("Failed to refresh content: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to refresh content",
		})
	}

	log.Printf("Successfully refreshed content from webhook")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "content refreshed successfully",
	})
}

// verifyWebhookSignature verifies that the webhook request came from GitHub
func verifyWebhookSignature(body []byte, signature, secret string) bool {
	if signature == "" {
		return false
	}

	// Remove "sha256=" prefix
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