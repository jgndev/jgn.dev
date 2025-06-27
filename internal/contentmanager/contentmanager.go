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

// retryWithBackoff executes a function with exponential backoff retry logic
func retryWithBackoff(fn func() error, maxRetries int, baseDelay time.Duration) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt < maxRetries {
				delay := baseDelay * time.Duration(1<<attempt) // exponential backoff: 1s, 2s, 4s
				log.Printf("Attempt %d failed, retrying in %v: %v", attempt+1, delay, err)
				time.Sleep(delay)
				continue
			}
		} else {
			return nil // success
		}
	}
	return lastErr
}

// ContentManager manages content storage and operations, providing synchronization and interaction with GitHub repositories.
type ContentManager struct {
	sync.RWMutex
	posts       map[string]Post
	client      *http.Client
	repoOwner   string
	repoName    string
	githubToken string
}

// NewContentManager initializes and returns a pointer to a new ContentManager instance for managing GitHub content.
// It requires the GitHub repository owner and name as input and retrieves the GITHUB_TOKEN from the environment.
func NewContentManager(repoOwner, repoName string) *ContentManager {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Println("Warning: GITHUB_TOKEN environment variable not set. API requests will be rate limited.")
	}

	return &ContentManager{
		posts:       make(map[string]Post),
		client:      &http.Client{},
		repoOwner:   repoOwner,
		repoName:    repoName,
		githubToken: githubToken,
	}
}

// listRepoContent retrieves the content of a GitHub repository for the provided path.
// It supports retrieving directories or single files and returns an array of githubContent items.
func (cm *ContentManager) listRepoContent(path string) ([]githubContent, error) {
	var contents []githubContent
	
	err := retryWithBackoff(func() error {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", cm.repoOwner, cm.repoName, path)

		log.Printf("fetching content from: %s", url)

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

// fetchFileContent retrieves the content of a file from a GitHub repository by its path.
// It decodes base64-encoded content if necessary and returns the file content or an error.
func (cm *ContentManager) fetchFileContent(path string) (string, error) {
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

// matchesAllTerms checks if all terms in the given list exist in the combined searchable fields of the provided post.
func matchesAllTerms(post Post, terms []string) bool {
	searchText := strings.ToLower(strings.Join([]string{
		post.Title,
		post.Summary,
		post.RawContent,
		strings.Join(post.Tags, " "),
	}, " "))

	for _, term := range terms {
		if !strings.Contains(searchText, term) {
			return false
		}
	}

	return true
}

// RefreshContent updates the internal state by fetching and parsing markdown files from a GitHub repository.
// It processes valid, published posts and skips ignored or non-markdown files.
// The method locks the manager for atomic updates and logs errors if any issues occur during processing.
func (cm *ContentManager) RefreshContent() error {
	// List files in the content directory
	files, err := cm.listRepoContent("")
	if err != nil {
		return fmt.Errorf("failed to list content: %v", err)
	}

	newPosts := make(map[string]Post)

	// Files to ignore
	ignoredFiles := map[string]bool{
		".gitignore": true,
		"README.md":  true,
		"LICENSE.md": true,
	}

	log.Printf("Found %d files in repository", len(files))

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

		log.Printf("Processing markdown file: %s", file.Name)

		content, err := cm.fetchFileContent(file.Path)
		if err != nil {
			log.Printf("Failed to fetch %s: %v", file.Name, err)
			return fmt.Errorf("failed to fecth %s: %w", file.Name, err)
		}

		post, err := parseMarkdown(content)
		if err != nil {
			log.Printf("Failed to parse %s: %v", file.Name, err)
			return fmt.Errorf("failed to parse %s: %w", file.Name, err)
		}

		log.Printf("Parsed post: Title='%s', Slug='%s', Published=%v, Tags=%v",
			post.Title, post.Slug, post.Published, post.Tags)

		// Check for empty slug
		if post.Slug == "" {
			log.Printf("WARNING: Post '%s' has empty slug, skipping", post.Title)
			continue
		}

		// Only include published posts
		if !post.Published {
			log.Printf("Skipping unpublished post: %s", post.Title)
			continue
		}

		newPosts[post.Slug] = post
	}

	log.Printf("Successfully processed %d posts", len(newPosts))

	// Update posts atomically
	cm.Lock()
	cm.posts = newPosts
	cm.Unlock()

	return nil
}

// GetAll retrieves all posts, sorts them by date in descending order, and returns them as a slice. It is thread-safe.
func (cm *ContentManager) GetAll() []Post {
	cm.RLock()
	defer cm.RUnlock()

	posts := make([]Post, 0, len(cm.posts))
	for _, post := range cm.posts {
		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	return posts
}

// GetByTag retrieves posts associated with a specific tag, sorted by date in descending order. It is thread-safe.
func (cm *ContentManager) GetByTag(tag string) []Post {
	cm.RLock()
	defer cm.RUnlock()

	var tagged []Post

	for _, post := range cm.posts {
		for _, t := range post.Tags {
			if t == tag {
				tagged = append(tagged, post)
				break
			}
		}
	}

	sort.Slice(tagged, func(i, j int) bool {
		return tagged[i].Date.After(tagged[j].Date)
	})

	return tagged
}

// GetRecent retrieves the most recent `n` posts sorted by date in descending order. Returns all posts if fewer than `n` exist.
func (cm *ContentManager) GetRecent(n int) []Post {
	posts := cm.GetAll()
	if len(posts) < n {
		return posts
	}

	return posts[:n]
}

// GetOldest retrieves the oldest `n` posts sorted by date in ascending order. Returns all posts if fewer than `n` exist.
func (cm *ContentManager) GetOldest(n int) []Post {
	posts := cm.GetAll()
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.Before(posts[j].Date)
	})

	if len(posts) < n {
		return posts
	}

	return posts[:n]
}

// Search filters posts by a given query string, returning all matches sorted by relevance and date in descending order.
func (cm *ContentManager) Search(query string) []Post {
	cm.RLock()
	defer cm.RUnlock()

	if query == "" {
		return []Post{}
	}

	terms := strings.Fields(strings.ToLower(query))
	matches := make(map[string]Post)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Process posts concurrently
	for slug, post := range cm.posts {
		wg.Add(1)
		go func(slug string, post Post) {
			defer wg.Done()

			if matchesAllTerms(post, terms) {
				mu.Lock()
				matches[slug] = post
				mu.Unlock()
			}
		}(slug, post)
	}
	wg.Wait()

	// Convert matches to slice
	results := make([]Post, 0, len(matches))
	for _, post := range matches {
		results = append(results, post)
	}

	// Sort by relevance, using date for now
	sort.Slice(results, func(i, j int) bool {
		return results[i].Date.After(results[j].Date)
	})

	return results
}

// GetBySlug retrieves a post by its slug from the content manager's storage. Returns the post and a boolean indicating existence.
func (cm *ContentManager) GetBySlug(slug string) (Post, bool) {
	cm.RLock()
	defer cm.RUnlock()

	post, exists := cm.posts[slug]
	return post, exists
}
