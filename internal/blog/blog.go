package blog

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/lib/pq"
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

// NewMyRenderer creates a new custom renderer.
// GetTailwindRenderer creates a new custom renderer instance.
func GetTailwindRenderer() *TailwindRenderer {
	// html.NewRenderer creates a standard renderer with common flags.
	return &TailwindRenderer{
		Renderer: html.NewRenderer(html.RendererOptions{
			Flags: html.CommonFlags,
		}),
	}
}

// RenderMarkdownToHTML converts a markdown string to an HTML string.
func RenderMarkdownToHTML(md string) string {
	renderer := GetTailwindRenderer()
	// The `markdown.ToHTML` function parses the markdown and returns the HTML as a byte slice.
	htmlBytes := markdown.ToHTML([]byte(md), nil, renderer)
	return string(htmlBytes)
}

type TailwindRenderer struct {
	*html.Renderer // This is the key part: struct embedding.
}

// SplitWithoutEmpty splits a string and filters out empty strings,
// returning the result as a pq.StringArray.
func SplitWithoutEmpty(s, sep string) pq.StringArray { // 2. Change the return type
	// 1. Split the string
	parts := strings.Split(s, sep)

	// 2. Filter out empty strings
	var result pq.StringArray // 3. Use pq.StringArray for the result variable
	for _, part := range parts {
		if part != "" {
			// Trim whitespace and append
			result = append(result, strings.TrimSpace(part))
		}
	}
	return result
}

// RenderNode implements the Renderer interface.
func (r *TailwindRenderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	switch node := node.(type) {
	case *ast.Heading:
		if node.Level == 1 {
			if entering {
				// Write the opening <h1> tag with Tailwind classes.
				w.Write([]byte(`<h1 class="text-4xl font-extrabold text-blue-800">`))
			} else {
				// Write the closing </h1> tag.
				w.Write([]byte(`</h1>`))
			}
			// Continue the traversal to the next node.
			return ast.GoToNext
		}
	case *ast.Text:
		// For text nodes, just write the content.
		w.Write(node.Literal)
		return ast.GoToNext
	case *ast.Paragraph:
		if entering {
			w.Write([]byte(`<p>`))
		} else {
			w.Write([]byte(`</p>`))
		}
		return ast.GoToNext
	case *ast.Image:
		if entering {
			// Cast the node to a *ast.Image to access its fields.
			image := node
			// Get the destination (URL) and title (alt text).
			dest := string(image.Destination)
			altText := string(image.Title)

			// Write the <img> tag with the source and alt attributes.
			// You can also add Tailwind classes here if desired.
			w.Write([]byte(fmt.Sprintf(`<img class="h-200 w-96 object-scale-down" src="%s" alt="%s" />`, dest, altText)))
		}
		// Since <img> is a self-closing tag, we don't need a separate "leaving" case.
		// We can just skip rendering the children (the alt text inside the markdown).
		return ast.SkipChildren
	}

	// For any other nodes not handled, just continue the traversal.
	return ast.GoToNext
}

// RenderHeader implements the Renderer interface.
// We don't need a specific header for this example, so it's a no-op.
func (r *TailwindRenderer) RenderHeader(w io.Writer, node ast.Node) {
	// No-op
}

var Categories = [7]string{"football", "cs", "politics", "plants", "cine", "lit", "random"}
