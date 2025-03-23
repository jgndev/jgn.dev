package contentmanager

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type ContentManager struct {
	sync.RWMutex
	posts         map[string]Post
	client        *http.Client
	repoOwner     string
	repoName      string
	githubToken   string
	WebhookSecret string
}

func NewContentManger(repoOwner, repoName, githubToken, webhookSecret string) *ContentManager {
	return &ContentManager{
		posts:         make(map[string]Post),
		client:        &http.Client{},
		repoOwner:     repoOwner,
		repoName:      repoName,
		githubToken:   githubToken,
		WebhookSecret: webhookSecret,
	}
}

func (cm *ContentManager) listRepoContent(path string) ([]githubContent, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
		cm.repoOwner, cm.repoName, path)

	log.Printf("fetching content from: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+cm.githubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := cm.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle single file response
	var singleContent githubContent
	if err := json.NewDecoder(resp.Body).Decode(&singleContent); err == nil {
		return []githubContent{singleContent}, nil
	}

	// Reset response body for array parsing
	resp.Body.Close()
	resp, _ = cm.client.Do(req)
	defer resp.Body.Close()

	var contents []githubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, err
	}

	return contents, nil
}

func (cm *ContentManager) fetchFileContent(path string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", cm.repoOwner, cm.repoName, path)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "token "+cm.githubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

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

func (cm *ContentManager) RefreshContent() error {
	// List files in content directory
	files, err := cm.listRepoContent("")
	if err != nil {
		return fmt.Errorf("failed to list content: %v", err)
	}

	newPosts := make(map[string]Post)

	// Process each markdown file
	for _, file := range files {
		if file.Type != "file" || !strings.HasSuffix(file.Name, ".md") {
			continue
		}

		content, err := cm.fetchFileContent(file.Path)
		if err != nil {
			return fmt.Errorf("failed to fecth %s: %w", file.Name, err)
		}

		post, err := parseMarkdown(content)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", file.Name, err)
		}

		newPosts[post.Slug] = post
	}

	log.Printf("found %d posts, files %v", len(newPosts), files)

	// Update posts atomically
	cm.Lock()
	cm.posts = newPosts
	cm.Unlock()

	return nil
}

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

func (cm *ContentManager) GetBySlug(slug string) (Post, bool) {
	cm.RLock()
	defer cm.RUnlock()
	post, exists := cm.posts[slug]
	return post, exists
}

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

func (cm *ContentManager) GetRecent(n int) []Post {
	posts := cm.GetAll()
	if len(posts) < n {
		return posts
	}

	return posts[:n]
}

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
