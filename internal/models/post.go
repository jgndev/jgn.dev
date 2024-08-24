package models

import (
	"errors"
	"google.golang.org/api/iterator"
	"log"
	"sort"
	"strings"
)

// Post is a struct of properties that make up a post in the application
type Post struct {
	ID          string   `firestore:"ID"`
	Title       string   `firestore:"title"`
	Date        string   `firestore:"date"`
	Author      string   `firestore:"author"`
	Summary     string   `firestore:"summary"`
	Slug        string   `firestore:"slug"`
	HTMLContent string   `firestore:"htmlContent"`
	Tags        []string `firestore:"tags"`
	Published   bool     `firestore:"published"`
}

// Posts is a collection of Post
type Posts []Post

// LoadFromFirestore retrieves all posts from the Firestore collection
func (p *Posts) LoadFromFirestore(ctx context.Context, client *firestore.Client) error {
	iter := client.Collection("posts").Documents(ctx)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return err
		}

		var post Post
		err = doc.DataTo(&post)
		if err != nil {
			return err
		}

		post.ID = doc.Ref.ID
		*p = append(*p, post) // Modify the original Blog slice
	}

	// Sort the posts newest to oldest
	sort.Slice(*p, func(i, j int) bool {
		return (*p)[i].Date > (*p)[j].Date
	})

	return nil
}

// GetAll retruns all posts
func (p Posts) GetAll() Posts {
	return p
}

// GetByID returns a specific post from the collection and a boolean value that
// represents a successful retrieval
func (p Posts) GetByID(id string) (*Post, bool) {
	for i := range p {
		if p[i].ID == id {
			return &p[i], true
		}
	}

	return nil, false
}

// GetBySlug returns a specific post by the value of the slug
func (p Posts) GetBySlug(slug string) (*Post, bool) {
	for i := range p {
		if p[i].Slug == slug {
			return &p[i], true
		}
	}

	return nil, false
}

// GetByTag returns a collection of posts that match the provided slug
func (p Posts) GetByTag(tag string) Posts {
	var taggedPosts Posts
	for _, post := range p {
		for _, t := range post.Tags {
			if t == tag {
				taggedPosts = append(taggedPosts, post)
				break
			}
		}
	}

	return taggedPosts
}

// GetRecent returns a collection of the n most recent posts
func (p Posts) GetRecent(n int) Posts {
	if len(p) < n {
		return p
	}

	return p[:n]
}

// GetOldest returns a collection of posts that are the n oldest
func (p Posts) GetOldest(n int) Posts {
	length := len(p)
	if len(p) < n {
		return p
	}

	return p[length-n:]
}

// FilterBySearchTerms returns a collection of posts that match a search term
func (p Posts) FilterBySearchTerms(query string) Posts {
	log.Println("query: ", query)

	var filteredPosts Posts
	uniqueMatches := mapset.NewSet[string]()

	// Split the query into individual search terms
	searchTerms := strings.Fields(query)
	log.Println("search terms: ", searchTerms)

	for _, s := range searchTerms {
		_ = strings.ToLower(s)
	}
	log.Println("search terms: ", searchTerms)

	for _, post := range p {
		matches := true

		for _, term := range searchTerms {
			lowerTitle := strings.ToLower(post.Title)
			lowerTerm := strings.ToLower(term)
			lowerSummary := strings.ToLower(post.Summary)
			if !strings.Contains(lowerTitle, lowerTerm) &&
				!strings.Contains(lowerSummary, lowerTerm) &&
				!contains(lowerTerm, post.Tags) {
				matches = false
				break
			}

			if matches && !uniqueMatches.Contains(post.ID) {
				uniqueMatches.Add(post.ID)
				filteredPosts = append(filteredPosts, post)
			}

		}
	}

	return filteredPosts
}

func contains(term string, tags []string) bool {
	for _, tag := range tags {
		if strings.EqualFold(term, tag) {
			return true
		}
	}

	return false
}
