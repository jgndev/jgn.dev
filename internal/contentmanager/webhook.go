package contentmanager

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io"
	"log"
)

type GitHubPushEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Commits []struct {
		Added    []string `json:"added"`
		Modified []string `json:"modified"`
		Removed  []string `json:"removed"`
	} `json:"commits"`
}

func WebhookHandler(contentManager *ContentManager) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Verify signature
		signature := c.Request().Header.Get("X-Hub-Signature-256")
		if signature == "" {
			return c.String(401, "No signature")
		}

		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(400, "Failed to read body")
		}

		if !verifySignature(body, signature, contentManager.WebhookSecret) {
			return c.String(401, "Invalid signature")
		}

		// Parse event
		var event GitHubPushEvent
		if err = json.Unmarshal(body, &event); err != nil {
			return c.String(400, "Invalid JSON")
		}

		// Only process main branch
		if event.Ref != "refs/head/main" {
			return c.NoContent(200)
		}

		// Async content refresh
		go func() {
			if err = contentManager.RefreshContent(); err != nil {
				log.Printf("Error refreshing content: %v", err)
			}
		}()

		return c.NoContent(200)
	}
}

func verifySignature(payload []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMac := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedMac))
}
