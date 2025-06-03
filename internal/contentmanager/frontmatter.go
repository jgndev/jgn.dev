package contentmanager

import "time"

// FrontMatter represents the metadata of a Markdown document, typically defined in the YAML front matter section.
type FrontMatter struct {
	ID        string    `yaml:"id"`
	Date      time.Time `yaml:"date"`
	Title     string    `yaml:"title"`
	Author    string    `yaml:"author"`
	Summary   string    `yaml:"summary"`
	Slug      string    `yaml:"slug"`
	Tags      []string  `yaml:"tags"`
	Published bool      `yaml:"published"`
}

// CheatsheetFrontMatter represents metadata for a cheatsheet, typically extracted from its frontmatter in YAML format.
type CheatsheetFrontMatter struct {
	ID        string    `yaml:"id"`
	Date      time.Time `yaml:"date"`
	Title     string    `yaml:"title"`
	Author    string    `yaml:"author"`
	Summary   string    `yaml:"summary"`
	Slug      string    `yaml:"slug"`
	Tags      []string  `yaml:"tags"`
	Published bool      `yaml:"published"`
}
