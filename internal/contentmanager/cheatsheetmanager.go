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
)

type CheatsheetManager struct {
	sync.RWMutex
	cheatsheets map[string]Cheatsheet
	client      *http.Client
	repoOwner   string
	repoName    string
	githubToken string
}

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

func (cm *CheatsheetManager) listRepoContent(path string) ([]githubContent, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", cm.repoOwner, cm.repoName, path)

	log.Printf("fetching cheatsheet content from: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Add authentication if token is available
	if cm.githubToken != "" {
		req.Header.Set("Authorization", "token "+cm.githubToken)
	}

	resp, err := cm.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// Try to decode as array first (directory listing)
	var contents []githubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		// If that fails, it might be a single file
		resp.Body.Close()
		resp, err = cm.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var singleContent githubContent
		if err := json.NewDecoder(resp.Body).Decode(&singleContent); err != nil {
			return nil, fmt.Errorf("failed to decode response as array or single file: %v", err)
		}
		return []githubContent{singleContent}, nil
	}

	return contents, nil
}

func (cm *CheatsheetManager) fetchFileContent(path string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", cm.repoOwner, cm.repoName, path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Add authentication if token is available
	if cm.githubToken != "" {
		req.Header.Set("Authorization", "token "+cm.githubToken)
	}

	resp, err := cm.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Encoding == "base64" {
		content, err := base64.StdEncoding.DecodeString(result.Content)
		if err != nil {
			return "", err
		}

		return string(content), nil
	}

	return result.Content, nil
}

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

func (cm *CheatsheetManager) RefreshContent() error {
	// List files in content directory
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

	// Process each markdown file
	for _, file := range files {
		// Skip if not a file or not a markdown file
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

		log.Printf("Parsed cheatsheet: Title='%s', Slug='%s', Published=%v, Tags=%v",
			cheatsheet.Title, cheatsheet.Slug, cheatsheet.Published, cheatsheet.Tags)

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

	log.Printf("Successfully processed %d cheatsheets", len(newCheatsheets))

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

func (cm *CheatsheetManager) GetRecent(n int) []Cheatsheet {
	cheatsheets := cm.GetAll()
	if len(cheatsheets) < n {
		return cheatsheets
	}

	return cheatsheets[:n]
}

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

func (cm *CheatsheetManager) GetBySlug(slug string) (Cheatsheet, bool) {
	cm.RLock()
	defer cm.RUnlock()

	cheatsheet, exists := cm.cheatsheets[slug]
	return cheatsheet, exists
}
