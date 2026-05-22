package blog

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/lib/pq"
	"github.com/microcosm-cc/bluemonday"
	"salada/internal/blog/model"
)

var (
	slugRegex    = regexp.MustCompile(`[^a-z0-9]+`)
	newlineRegex = regexp.MustCompile("(?s)```.*?```|`.*?`|(\n\n)")
	linkRegex    = regexp.MustCompile(`!?\[.*?\]\(.*?\)`)
	imageRegex   = regexp.MustCompile(`!\[.*?\]\((.*?)\)`)
)

// CreateSlug generates a URL-friendly slug from a string.
func CreateSlug(s string) string {
	s = strings.ToLower(s)
	s = slugRegex.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// GetTailwindRenderer returns a new Tailwind-styled HTML renderer.
func GetTailwindRenderer() *TailwindRenderer {
	return &TailwindRenderer{
		Renderer: html.NewRenderer(html.RendererOptions{
			Flags: html.CommonFlags,
		}),
	}
}

// extractTextFromNode gets plain text from an AST node.
func extractTextFromNode(node ast.Node) string {
	var text strings.Builder
	ast.WalkFunc(node, func(n ast.Node, entering bool) ast.WalkStatus {
		if entering {
			if t, ok := n.(*ast.Text); ok {
				text.Write(t.Literal)
			}
		}
		return ast.GoToNext
	})
	return strings.TrimSpace(text.String())
}

// generateHeadingID creates a URL-friendly ID for headings.
func generateHeadingID(text string) string {
	return CreateSlug(text)
}

// RenderMarkdownToHTML converts markdown to sanitized HTML.
func RenderMarkdownToHTML(md string) string {
	html, _ := RenderMarkdownToHTMLWithIDs(md)
	return html
}

// RenderMarkdownToHTMLWithIDs converts markdown to HTML and extracts the Table of Contents.
func RenderMarkdownToHTMLWithIDs(md string) (string, []model.TocItem) {
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs | parser.HardLineBreak)

	// Preserve visual line breaks without creating paragraphs
	md = newlineRegex.ReplaceAllStringFunc(md, func(match string) string {
		if match == "\n\n" {
			return "\n\u200B\n"
		}
		return match
	})

	doc := p.Parse([]byte(md))

	var tocItems []model.TocItem
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			if heading, ok := node.(*ast.Heading); ok && heading.Level >= 1 && heading.Level <= 3 {
				text := extractTextFromNode(heading)
				if text != "" {
					tocItems = append(tocItems, model.TocItem{
						Level: heading.Level,
						Text:  text,
						ID:    generateHeadingID(text),
					})
				}
			}
		}
		return ast.GoToNext
	})

	renderer := GetTailwindRenderer()
	htmlBytes := markdown.Render(doc, renderer)

	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("class", "id").Globally()
	policy.AllowElements("button")
	policy.AllowRelativeURLs(true)
	policy.AllowDataURIImages()

	return string(policy.SanitizeBytes(htmlBytes)), tocItems
}

type TailwindRenderer struct {
	*html.Renderer
}

// SplitWithoutEmpty splits a string and returns a non-empty trimmed pq.StringArray.
func SplitWithoutEmpty(s, sep string) pq.StringArray {
	parts := strings.Split(s, sep)
	result := make(pq.StringArray, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// RenderNode implements the markdown.Renderer interface for Tailwind styling.
func (r *TailwindRenderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	switch node := node.(type) {
	case *ast.Heading:
		idAttr := ""
		if id := string(node.HeadingID); id != "" {
			idAttr = fmt.Sprintf(` id="%s"`, id)
		}

		if entering {
			class := ""
			switch node.Level {
			case 1:
				class = "text-4xl font-extrabold text-blue-800"
			case 2:
				class = "text-3xl font-bold text-blue-800 mt-8 mb-4"
			case 3:
				class = "text-2xl font-semibold text-blue-800 mt-6 mb-3"
			default:
				class = "text-xl font-semibold text-blue-800 mt-4 mb-2"
			}
			fmt.Fprintf(w, `<h%d class="%s"%s>`, node.Level, class, idAttr)
		} else {
			fmt.Fprintf(w, `</h%d>`, node.Level)
		}

	case *ast.Strong:
		if entering {
			w.Write([]byte("<strong>"))
		} else {
			w.Write([]byte("</strong>"))
		}

	case *ast.Text:
		content := string(node.Literal)
		if strings.Contains(content, "\u200B") {
			parts := strings.Split(content, "\u200B")
			for i, part := range parts {
				html.EscapeHTML(w, []byte(part))
				if i < len(parts)-1 {
					w.Write([]byte(`<span style="user-select: none;">&#8203;</span>`))
				}
			}
		} else {
			html.EscapeHTML(w, node.Literal)
		}

	case *ast.Paragraph:
		if entering {
			w.Write([]byte(`<p>`))
		} else {
			w.Write([]byte(`</p>`))
		}

	case *ast.Image:
		if entering {
			fmt.Fprintf(w, `<img class="w-full h-auto" src="%s" alt="%s" />`, node.Destination, node.Title)
		}
		return ast.SkipChildren

	case *ast.Link:
		if entering {
			fmt.Fprintf(w, `<a href="%s" class="text-blue-600 hover:text-blue-800 transition duration-150">`, node.Destination)
		} else {
			w.Write([]byte(`</a>`))
		}

	case *ast.CodeBlock:
		w.Write([]byte(`<div class="relative group">`))
		w.Write([]byte(`<button class="copy-button absolute top-2 right-2 px-2 py-1 text-xs font-sans text-gray-300 bg-gray-700 hover:bg-gray-600 rounded opacity-0 group-hover:opacity-100 transition-opacity duration-200 focus:outline-none focus:ring-1 focus:ring-gray-500">Copy</button>`))

		langClass := ""
		if len(node.Info) > 0 {
			langClass = fmt.Sprintf(" language-%s", strings.Fields(string(node.Info))[0])
		}

		fmt.Fprintf(w, `<pre class="bg-gray-800 text-white p-4 rounded-lg overflow-x-auto my-4 text-sm font-mono shadow-inner"><code class="block%s">`, langClass)
		clean := strings.ReplaceAll(string(node.Literal), "\u200B", "")
		html.EscapeHTML(w, []byte(clean))
		w.Write([]byte(`</code></pre></div>`))
		return ast.SkipChildren

	case *ast.Code:
		w.Write([]byte(`<code class="bg-gray-100 text-red-600 px-1 rounded">`))
		clean := strings.ReplaceAll(string(node.Literal), "\u200B", "")
		html.EscapeHTML(w, []byte(clean))
		w.Write([]byte(`</code>`))
		return ast.SkipChildren

	case *ast.Hardbreak:
		w.Write([]byte(`<br>`))
	}
	return ast.GoToNext
}

// RenderHeader is a no-op implementation of the Renderer interface.
func (r *TailwindRenderer) RenderHeader(w io.Writer, node ast.Node) {}

// DeleteFiles removes files at the given paths and returns any errors.
func DeleteFiles(filepaths []string) []string {
	var errs []string
	for _, path := range filepaths {
		if err := os.Remove("." + path); err != nil {
			errs = append(errs, fmt.Sprintf("failed to delete %s: %v", path, err))
		}
	}
	return errs
}

// stripMarkdownLinks removes markdown link and image syntax from a string.
func stripMarkdownLinks(s string) string {
	s = linkRegex.ReplaceAllString(s, "")
	return strings.ReplaceAll(s, "  ", " ")
}

// GetContentSummary returns a truncated version of the content after stripping markdown links.
func GetContentSummary(s string, maxChars int) string {
	s = stripMarkdownLinks(s)
	if utf8.RuneCountInString(s) <= maxChars {
		return s
	}

	var count int
	for i := range s {
		if count == maxChars {
			return s[:i] + "..."
		}
		count++
	}
	return s
}

// GetFirstImage returns the URL of the first image found in the markdown content.
func GetFirstImage(content string) string {
	match := imageRegex.FindStringSubmatch(content)
	if len(match) > 1 {
		url := match[1]
		if idx := strings.Index(url, " \""); idx != -1 {
			return url[:idx]
		}
		return url
	}
	return ""
}
