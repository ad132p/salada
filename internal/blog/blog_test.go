package blog

import (
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRenderInlineCode(t *testing.T) {
	md := "This is `inline code` and this is text after."
	html := RenderMarkdownToHTML(md)
	
	// Check if the code tag is closed correctly
	assert.Contains(t, html, `<code class="bg-gray-100 text-red-600 px-1 rounded">inline code</code>`)
	assert.Contains(t, html, "and this is text after.")
	
	// Check if styles are leaking (this is a bit subjective but we can check if the tag is closed)
	// If the tag is not closed, the rest of the text might be inside it or follow it in a way that's broken.
	// Bluemonday might "fix" it by closing it at the end of the document, which explains why styles keep being applied.
	
	countOpen := strings.Count(html, "<code")
	countClose := strings.Count(html, "</code>")
	assert.Equal(t, countOpen, countClose, "Number of opening code tags should match closing tags")
}

func TestRenderHardbreak(t *testing.T) {
	// In some markdown flavors, two spaces at the end of a line or a backslash results in a hard break.
	// RenderMarkdownToHTML uses parser.HardLineBreak extension.
	md := "Line 1  \nLine 2"
	html := RenderMarkdownToHTML(md)
	
	// Count <br> tags. It should be exactly 1.
	count := strings.Count(html, "<br>")
	assert.Equal(t, 1, count, "Should have exactly one <br> tag")
}
