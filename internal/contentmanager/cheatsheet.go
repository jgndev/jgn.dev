package contentmanager

import "time"

// Cheatsheet represents a structured data model for storing information about a cheatsheet, including metadata and content.
type Cheatsheet struct {
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
