package contentmanager

import (
	"bytes"
	"github.com/spf13/viper"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"regexp"
	"strings"
	"time"
)

func parseMarkdown(content string) (Post, error) {
	frontMatter, body, err := parseFrontMatter([]byte(content))
	if err != nil {
		return Post{}, err
	}

	//style := styles.Get("monokai")
	output, err := markdownToHtml([]byte(body))
	if err != nil {
		return Post{}, err
	}

	return Post{
		ID:          frontMatter.ID,
		Date:        frontMatter.Date,
		DisplayDate: frontMatter.Date.Format(time.RFC3339),
		Title:       frontMatter.Title,
		Content:     output,
		RawContent:  body,
		Slug:        frontMatter.Slug,
		Tags:        frontMatter.Tags,
		Published:   frontMatter.Published,
	}, nil
}

func parseFrontMatter(markdown []byte) (FrontMatter, string, error) {
	parts := strings.SplitN(string(markdown), "---", 3)
	if len(parts) < 3 {
		return FrontMatter{}, string(markdown), nil
	}

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewBufferString(parts[1])); err != nil {
		return FrontMatter{}, "", err
	}

	var fm FrontMatter
	if err := v.Unmarshal(&fm); err != nil {
		return FrontMatter{}, "", err
	}

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

	// Fix code classes for highlight.js
	codePattern := regexp.MustCompile(`<pre><code class="language-(\w+)">`)
	output = codePattern.ReplaceAllString(output, `<pre><code class="$1">`)

	return output, nil
}
