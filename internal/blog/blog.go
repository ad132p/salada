package blog

import (
	"regexp"
	//"salada/internal/blog/model"
	"strings"
	//"golang.org/x/crypto/bcrypt"
	"github.com/gomarkdown/markdown"
)

// createSlug generates a URL-friendly slug from a given title.
func CreateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace non-alphanumeric characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}

// RenderMarkdownToHTML converts a markdown string to an HTML string.
func RenderMarkdownToHTML(md string) string {
	// The `markdown.ToHTML` function parses the markdown and returns the HTML as a byte slice.
	htmlBytes := markdown.ToHTML([]byte(md), nil, nil)
	return string(htmlBytes)
}
