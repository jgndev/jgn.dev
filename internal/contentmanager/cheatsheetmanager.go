package contentmanager

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// CheatsheetManager manages the retrieval, storage, and filtering of cheatsheets from a remote GitHub repository.
type CheatsheetManager struct {
	sync.RWMutex
	cheatsheets map[string]Cheatsheet
	client      *http.Client
	repoOwner   string
	repoName    string
	githubToken string
}

// NewCheatsheetManager initializes and returns a new instance of CheatsheetManager with the given repository owner and name.
func NewCheatsheetManager(repoOwner, repoName string) *CheatsheetManager {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Println("Warning: GITHUB_TOKEN environment variable not set. API requests will be rate limited.")
	}

	return &CheatsheetManager{
		cheatsheets: make(map[string]Cheatsheet),
		client:      &http.Client{},
		repoOwner:   repoOwner,
		repoName:    repoName,
		githubToken: githubToken,
	}
}

// listRepoContent retrieves the content of a GitHub repository at a specified path using the GitHub API.
// It handles both directory listings and single file retrieval, returning a slice of githubContent or an error.
func (cm *CheatsheetManager) listRepoContent(path string) ([]githubContent, error) {
	var contents []githubContent

	err := retryWithBackoff(func() error {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", cm.repoOwner, cm.repoName, path)

		log.Printf("fetching cheatsheet content from: %s", url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")

		// Add authentication if a token is available
		if cm.githubToken != "" {
			req.Header.Set("Authorization", "token "+cm.githubToken)
		}

		resp, err := cm.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
		}

		// Try to decode as an array first (directory listing)
		if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
			// If that fails, it might be a single file
			resp.Body.Close()
			resp, err = cm.client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var singleContent githubContent
			if err := json.NewDecoder(resp.Body).Decode(&singleContent); err != nil {
				return fmt.Errorf("failed to decode response as array or single file: %v", err)
			}
			contents = []githubContent{singleContent}
		}

		return nil
	}, 3, time.Second)

	return contents, err
}

// fetchFileContent retrieves the content of a file from a GitHub repository using the GitHub API.
// It decodes the content if it is encoded in base64 and returns it as a string. Returns an error if the operation fails.
func (cm *CheatsheetManager) fetchFileContent(path string) (string, error) {
	var content string

	err := retryWithBackoff(func() error {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", cm.repoOwner, cm.repoName, path)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")

		// Add authentication if a token is available
		if cm.githubToken != "" {
			req.Header.Set("Authorization", "token "+cm.githubToken)
		}

		resp, err := cm.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var result struct {
			Content  string `json:"content"`
			Encoding string `json:"encoding"`
		}

		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		if result.Encoding == "base64" {
			decoded, err := base64.StdEncoding.DecodeString(result.Content)
			if err != nil {
				return err
			}
			content = string(decoded)
		} else {
			content = result.Content
		}

		return nil
	}, 3, time.Second)

	return content, err
}

// matchesAllTermsCheatsheet checks if all the provided search terms are present in a Cheatsheet's combined text fields.
func matchesAllTermsCheatsheet(cheatsheet Cheatsheet, terms []string) bool {
	searchText := strings.ToLower(strings.Join([]string{
		cheatsheet.Title,
		cheatsheet.Summary,
		cheatsheet.RawContent,
		strings.Join(cheatsheet.Tags, " "),
	}, " "))

	for _, term := range terms {
		if !strings.Contains(searchText, term) {
			return false
		}
	}

	return true
}

// RefreshContent synchronizes the cheatsheet repository and updates the in-memory map with published Markdown files.
func (cm *CheatsheetManager) RefreshContent() error {
	// List files in the content directory
	files, err := cm.listRepoContent("")
	if err != nil {
		return fmt.Errorf("failed to list cheatsheet content: %v", err)
	}

	newCheatsheets := make(map[string]Cheatsheet)

	// Files to ignore
	ignoredFiles := map[string]bool{
		".gitignore": true,
		"README.md":  true,
		"LICENSE.md": true,
	}

	log.Printf("Found %d files in cheatsheets repository", len(files))

	// Process each Markdown file
	for _, file := range files {
		// Skip if not a file or not a Markdown file
		if file.Type != "file" || !strings.HasSuffix(file.Name, ".md") {
			log.Printf("Skipping non-markdown file: %s (type: %s)", file.Name, file.Type)
			continue
		}

		// Skip ignored files
		if ignoredFiles[file.Name] {
			log.Printf("Skipped ignored file: %s", file.Name)
			continue
		}

		log.Printf("Processing cheatsheet markdown file: %s", file.Name)

		content, err := cm.fetchFileContent(file.Path)
		if err != nil {
			log.Printf("Failed to fetch %s: %v", file.Name, err)
			return fmt.Errorf("failed to fetch %s: %w", file.Name, err)
		}

		cheatsheet, err := parseCheatsheetMarkdown(content)
		if err != nil {
			log.Printf("Failed to parse %s: %v", file.Name, err)
			return fmt.Errorf("failed to parse %s: %w", file.Name, err)
		}

		// Check for empty slug
		if cheatsheet.Slug == "" {
			log.Printf("WARNING: Cheatsheet '%s' has empty slug, skipping", cheatsheet.Title)
			continue
		}

		// Only include published cheatsheets
		if !cheatsheet.Published {
			log.Printf("Skipping unpublished cheatsheet: %s", cheatsheet.Title)
			continue
		}

		newCheatsheets[cheatsheet.Slug] = cheatsheet
	}

	// Update cheatsheets atomically
	cm.Lock()
	cm.cheatsheets = newCheatsheets
	cm.Unlock()

	return nil
}

func (cm *CheatsheetManager) GetAll() []Cheatsheet {
	cm.RLock()
	defer cm.RUnlock()

	cheatsheets := make([]Cheatsheet, 0, len(cm.cheatsheets))
	for _, cheatsheet := range cm.cheatsheets {
		cheatsheets = append(cheatsheets, cheatsheet)
	}

	sort.Slice(cheatsheets, func(i, j int) bool {
		return cheatsheets[i].Date.After(cheatsheets[j].Date)
	})

	return cheatsheets
}

// GetByTag retrieves all cheatsheets associated with the specified tag, sorted by date in descending order.
func (cm *CheatsheetManager) GetByTag(tag string) []Cheatsheet {
	cm.RLock()
	defer cm.RUnlock()

	var tagged []Cheatsheet

	for _, cheatsheet := range cm.cheatsheets {
		for _, t := range cheatsheet.Tags {
			if t == tag {
				tagged = append(tagged, cheatsheet)
				break
			}
		}
	}

	sort.Slice(tagged, func(i, j int) bool {
		return tagged[i].Date.After(tagged[j].Date)
	})

	return tagged
}

// GetRecent retrieves the most recent n cheatsheets, sorted by date in descending order. Returns all cheatsheets if n exceeds the total count.
func (cm *CheatsheetManager) GetRecent(n int) []Cheatsheet {
	cheatsheets := cm.GetAll()
	if len(cheatsheets) < n {
		return cheatsheets
	}

	return cheatsheets[:n]
}

// Search performs a case-insensitive search across all cheatsheets, returning those that match the query terms.
// The results are sorted by relevance, currently determined by the cheatsheet's date in descending order.
func (cm *CheatsheetManager) Search(query string) []Cheatsheet {
	cm.RLock()
	defer cm.RUnlock()

	if query == "" {
		return []Cheatsheet{}
	}

	terms := strings.Fields(strings.ToLower(query))
	matches := make(map[string]Cheatsheet)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Process cheatsheets concurrently
	for slug, cheatsheet := range cm.cheatsheets {
		wg.Add(1)
		go func(slug string, cheatsheet Cheatsheet) {
			defer wg.Done()

			if matchesAllTermsCheatsheet(cheatsheet, terms) {
				mu.Lock()
				matches[slug] = cheatsheet
				mu.Unlock()
			}
		}(slug, cheatsheet)
	}
	wg.Wait()

	// Convert matches to slice
	results := make([]Cheatsheet, 0, len(matches))
	for _, cheatsheet := range matches {
		results = append(results, cheatsheet)
	}

	// Sort by relevance, using date for now
	sort.Slice(results, func(i, j int) bool {
		return results[i].Date.After(results[j].Date)
	})

	return results
}

// GetBySlug retrieves a cheatsheet by its unique slug. Returns the Cheatsheet and a boolean indicating its existence.
func (cm *CheatsheetManager) GetBySlug(slug string) (Cheatsheet, bool) {
	cm.RLock()
	defer cm.RUnlock()

	cheatsheet, exists := cm.cheatsheets[slug]
	return cheatsheet, exists
}
