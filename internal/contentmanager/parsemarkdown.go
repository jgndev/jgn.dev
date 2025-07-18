package contentmanager

import (
	"bytes"
	"log"

	"github.com/spf13/viper"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"

	// "regexp"
	"strings"
	"time"
)

// parseMarkdown parses a Markdown string into a Post struct, extracting front matter and converting content to HTML.
func parseMarkdown(content string) (Post, error) {
	fm, body, err := parseFrontMatter([]byte(content))
	if err != nil {
		return Post{}, err
	}

	output, err := markdownToHtml([]byte(body))
	if err != nil {
		return Post{}, err
	}

	return Post{
		ID:          fm.ID,
		Date:        fm.Date,
		DisplayDate: fm.Date.Format(time.RFC3339),
		Title:       fm.Title,
		Author:      fm.Author,
		Summary:     fm.Summary,
		Content:     output,
		RawContent:  body,
		Slug:        fm.Slug,
		Tags:        fm.Tags,
		Published:   fm.Published,
	}, nil
}

// parseCheatsheetMarkdown parses a Markdown string into a Cheatsheet struct by extracting frontmatter and converting content to HTML.
// Returns the parsed Cheatsheet or an error if parsing or conversion fails.
func parseCheatsheetMarkdown(content string) (Cheatsheet, error) {
	fm, body, err := parseCheatsheetFrontMatter([]byte(content))
	if err != nil {
		return Cheatsheet{}, err
	}

	output, err := markdownToHtml([]byte(body))
	if err != nil {
		return Cheatsheet{}, err
	}

	return Cheatsheet{
		ID:          fm.ID,
		Date:        fm.Date,
		DisplayDate: fm.Date.Format(time.RFC3339),
		Title:       fm.Title,
		Author:      fm.Author,
		Summary:     fm.Summary,
		Content:     output,
		RawContent:  body,
		Slug:        fm.Slug,
		Tags:        fm.Tags,
		Published:   fm.Published,
	}, nil
}

// parseFrontMatter parses the front matter and content from a Markdown file, returning the front matter, body, and any errors.
// Front matter must be YAML and enclosed by `---` separators. If parsing fails, an error is returned.
func parseFrontMatter(markdown []byte) (FrontMatter, string, error) {
	parts := strings.SplitN(string(markdown), "---", 3)
	if len(parts) < 3 {
		log.Printf("No frontmatter found in markdown content (length: %d)", len(markdown))
		return FrontMatter{}, string(markdown), nil
	}

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(parts[1])); err != nil {
		log.Printf("Failed to parse YAML frontmatter: %v", err)
		return FrontMatter{}, "", err
	}

	var fm FrontMatter
	if err := v.Unmarshal(&fm); err != nil {
		log.Printf("Failed to unmarshal frontmatter into struct: %v", err)
		return FrontMatter{}, "", err
	}

	return fm, parts[2], nil
}

// parseCheatsheetFrontMatter extracts and parses the frontmatter YAML from cheatsheet Markdown content.
// Returns the parsed frontmatter, the remaining Markdown content, and an error if parsing fails.
func parseCheatsheetFrontMatter(markdown []byte) (CheatsheetFrontMatter, string, error) {
	parts := strings.SplitN(string(markdown), "---", 3)
	if len(parts) < 3 {
		log.Printf("No frontmatter found in cheatsheet markdown content (length: %d)", len(markdown))
		return CheatsheetFrontMatter{}, string(markdown), nil
	}

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(parts[1])); err != nil {
		log.Printf("Failed to parse YAML cheatsheet frontmatter: %v", err)
		return CheatsheetFrontMatter{}, "", err
	}

	var fm CheatsheetFrontMatter
	if err := v.Unmarshal(&fm); err != nil {
		log.Printf("Failed to unmarshal cheatsheet frontmatter into struct: %v", err)
		return CheatsheetFrontMatter{}, "", err
	}

	return fm, parts[2], nil
}

// markdownToHtml converts a Markdown input to HTML, using Goldmark with extensions like GFM, Linkify, and unsafe rendering.
// Returns the converted HTML string or an error if the conversion process fails.
func markdownToHtml(markdown []byte) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			extension.Linkify,
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(markdown, &buf); err != nil {
		return "", err
	}

	output := buf.String()

	// Replace target="_blank" for links
	output = strings.ReplaceAll(output, `<a href=`, `<a target="_blank" href=`)

	// Note: Keep the original language-* classes for highlight.js
	// Highlight.js expects: <code class="language-go">
	// No modification needed since Goldmark already outputs this format

	return output, nil
}
