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
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.HardLineBreak
	p := parser.NewWithExtensions(extensions)

	renderer := GetTailwindRenderer()

	// Pre-process markdown to handle empty newlines
	// We replace double newlines with a newline + zero-width space + newline
	// This prevents markdown from creating new paragraphs while maintaining the visual line break
	for strings.Contains(md, "\n\n") {
		md = strings.ReplaceAll(md, "\n\n", "\n\u200B\n")
	}

	// Parse the markdown using the configured parser
	doc := p.Parse([]byte(md))

	// Create the HTML renderer
	htmlBytes := markdown.Render(doc, renderer)

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
		html.EscapeHTML(w, node.Literal)
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
			w.Write([]byte(fmt.Sprintf(`<img class="w-96 h-auto object-scale-down" src="%s" alt="%s" />`, dest, altText)))
		}
		// Since <img> is a self-closing tag, we don't need a separate "leaving" case.
		// We can just skip rendering the children (the alt text inside the markdown).
		return ast.SkipChildren
	case *ast.Link:
		if entering {
			// Cast the node to *ast.Link to access its fields.
			link := node
			dest := string(link.Destination)

			// Write the opening <a> tag with href and Tailwind classes.
			// You'll want to choose appropriate Tailwind classes for a link (e.g., text-blue-600 hover:underline).
			w.Write([]byte(fmt.Sprintf(`<a href="%s" class="text-blue-600 hover:text-blue-800 transition duration-150">`, dest)))
		} else {
			// Write the closing </a> tag.
			w.Write([]byte(`</a>`))
		}
		return ast.GoToNext
	case *ast.CodeBlock:
		// Render code blocks with Tailwind styling
		w.Write([]byte(`<pre class="bg-gray-800 text-white p-4 rounded-lg overflow-x-auto my-4 text-sm font-mono shadow-inner"><code class="block">`))
		// Escape HTML characters in the code content
		html.EscapeHTML(w, node.Literal)
		w.Write([]byte(`</code></pre>`))
		return ast.SkipChildren
	case *ast.Hardbreak:
		w.Write([]byte(`<br>`))
		return ast.GoToNext
	}

	// For any other nodes not handled, just continue the traversal.
	return ast.GoToNext
}

// RenderHeader implements the Renderer interface.
// We don't need a specific header for this example, so it's a no-op.
func (r *TailwindRenderer) RenderHeader(w io.Writer, node ast.Node) {
	// No-op
}

// DeleteFiles takes a slice of strings (file paths) and attempts to delete
// the file at each path. It returns a slice of error strings containing
// details for any paths that could not be deleted.
func DeleteFiles(filepaths []string) []string {
	var deletionErrors []string

	for _, path := range filepaths {
		// os.Remove deletes the named file or directory.
		// If the path is a directory, it must be empty.
		err := os.Remove("." + path)

		if err != nil {
			// If an error occurs (e.g., file not found, permission denied),
			// format the error message and append it to the error slice.
			errMsg := fmt.Sprintf("Error deleting file '%s': %v", path, err)
			deletionErrors = append(deletionErrors, errMsg)
		} else {
			// Optional: Print a success message for files that were deleted.
			fmt.Printf("Successfully deleted: %s\n", path)
		}
	}

	return deletionErrors
}

// stripMarkdownLinks replaces any Markdown links or image references: [text](url) or ![text](url).
func stripMarkdownLinks(content string) string {
	// UPDATED REGEX: !? matches an optional exclamation mark.
	// This catches both images (e.g., ![alt](url)) and plain links (e.g., [text](url)).
	re := regexp.MustCompile(`!?\[.*?\]\(.*?\)`)

	// Replace all matches with an empty string.
	cleanedContent := re.ReplaceAllString(content, "")

	// Optional: Clean up any double spaces that might result from the replacement
	cleanedContent = strings.ReplaceAll(cleanedContent, "  ", " ")

	return cleanedContent
}

// GetContentSummary extracts the first N characters of a post after stripping
// sensitive/structural Markdown, ensuring it handles multi-byte UTF-8 characters correctly.
func GetContentSummary(fullContent string, maxChars int) string {
	// 1. Strip the file references (links/images)
	cleanedContent := stripMarkdownLinks(fullContent)

	// 2. Safely truncate the string based on runes (characters), not bytes.
	if utf8.RuneCountInString(cleanedContent) <= maxChars {
		return cleanedContent
	}

	// Iterate over the string as runes (characters)
	runes := []rune(cleanedContent)

	// Truncate to the desired length
	truncated := string(runes[:maxChars])

	// Add an ellipsis (...) for visual indication of truncation
	return truncated + "..."
}

var Categories = [7]string{"football", "cs", "politics", "plants", "cine", "lit", "random"}

// GetFirstImage extracts the URL of the first image found in the markdown content.
func GetFirstImage(content string) string {
	// Regex to match markdown images: ![alt](url)
	// Captures the URL in the first group.
	// We use a non-greedy match for the URL part to handle potential titles or closing parens.
	// The previous regex `([^)\s"]+)` stopped at spaces, which broke URLs with spaces.
	// This new regex `\(([^)]+)\)` captures everything inside the parentheses.
	// However, markdown links can have titles: `(url "title")`.
	// So we should capture until the first space OR closing paren if no space.
	// But wait, if the URL has spaces (not encoded), standard markdown might break,
	// but the user says they are rendering it, so it might be valid in their context.
	// Let's try to capture everything inside `(` and `)` and then strip potential title?
	// Or just be more permissive.
	//
	// If the input is `![alt](/uploads/foo bar.jpg)`, we want `/uploads/foo bar.jpg`.
	// If the input is `![alt](/uploads/foo.jpg "title")`, we want `/uploads/foo.jpg`.
	//
	// Let's use a regex that matches `(` then captures until `)` or ` "`.
	// actually, `!\[.*?\]\((.*?)\)` captures everything.
	// Then we can trim the title if present.
	re := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)
	match := re.FindStringSubmatch(content)
	if len(match) > 1 {
		url := match[1]
		// If there's a title (e.g. `url "title"`), strip it.
		// We split by space followed by a quote, or just the first space if we assume URLs don't have spaces...
		// BUT the user explicitly said "strings with spaces are not being handled", implying the URL HAS spaces.
		// Markdown standard says URLs with spaces should be encoded or wrapped in <>.
		// But if the user has raw spaces, we should probably take the whole thing?
		//
		// If we have `foo bar.jpg "title"`, splitting by space breaks the URL.
		// If we have `foo bar.jpg`, splitting by space breaks the URL.
		//
		// If the user is using standard markdown, URLs with spaces MUST be encoded (%20).
		// If they are NOT encoded, it's technically invalid markdown, but some parsers handle it.
		//
		// However, if the user says "it only shows 'WhatsApp' string", it means my previous regex `([^)\s"]+)` stopped at the space.
		//
		// If I change it to `(.*?)`, it will capture `WhatsApp Image... .jpeg`.
		// But what if there is a title? `WhatsApp Image... .jpeg "My Title"`.
		// Then I capture the title too.
		//
		// A common heuristic: if there is a quote `"`, the URL ends before it.
		// Let's try to capture everything, then check for quotes.

		// Check for title starting with "
		if idx := strings.Index(url, " \""); idx != -1 {
			return url[:idx]
		}
		return url
	}
	return ""
}
