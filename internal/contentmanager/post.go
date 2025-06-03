package contentmanager

import "time"

// Post represents a blog post with metadata and content information. It includes details like title, author, and tags.
type Post struct {
	ID          string    `yaml:"ID"`
	Date        time.Time `yaml:"date"`
	DisplayDate string
	Title       string `yaml:"title"`
	Author      string `yaml:"author"`
	Summary     string `yaml:"summary"`
	Content     string
	RawContent  string
	Slug        string   `yaml:"slug"`
	Tags        []string `yaml:"tags"`
	Published   bool     `yaml:"published"`
}
