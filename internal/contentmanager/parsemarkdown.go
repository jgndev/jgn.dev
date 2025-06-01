package contentmanager

import (
	"bytes"
	"github.com/spf13/viper"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"log"
	// "regexp"
	"strings"
	"time"
)

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
		Content:     output,
		RawContent:  body,
		Slug:        fm.Slug,
		Tags:        fm.Tags,
		Published:   fm.Published,
	}, nil
}

func parseFrontMatter(markdown []byte) (FrontMatter, string, error) {
	parts := strings.SplitN(string(markdown), "---", 3)
	if len(parts) < 3 {
		log.Printf("No frontmatter found in markdown content (length: %d)", len(markdown))
		return FrontMatter{}, string(markdown), nil
	}

	log.Printf("Parsing frontmatter: %s", strings.TrimSpace(parts[1]))

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

	log.Printf("Successfully parsed frontmatter: Title='%s', Slug='%s', Published=%v", 
		fm.Title, fm.Slug, fm.Published)

	return fm, parts[2], nil
}

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

	output := string(buf.Bytes())

	// Replace target="_blank" for links
	output = strings.ReplaceAll(output, `<a href=`, `<a target="_blank" href=`)

	// Note: Keep the original language-* classes for highlight.js
	// Highlight.js expects: <code class="language-go">
	// No modification needed since Goldmark already outputs this format

	return output, nil
}
