package models

// FrontMatter is the model for frontmatter in a markdown post read from Cloud Storage
type FrontMatter struct {
	Date      string   `yaml:"date"`
	ID        string   `yaml:"ID"`
	Title     string   `yaml:"title"`
	Author    string   `yaml:"author"`
	Summary   string   `yaml:"summary"`
	Slug      string   `yaml:"slug"`
	Tags      []string `yaml:"tags"`
	Published bool     `yaml:"published"`
}
