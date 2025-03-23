package contentmanager

import "time"

type FrontMatter struct {
	ID        string    `yaml:"ID"`
	Date      time.Time `yaml:"date"`
	Title     string    `yaml:"title"`
	Author    string    `yaml:"author"`
	Summary   string    `yaml:"summary"`
	Slug      string    `yaml:"slug"`
	Tags      []string  `yaml:"tags"`
	Published bool      `yaml:"published"`
}
