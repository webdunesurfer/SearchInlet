package optimizer

import (
	"strings"
	"testing"
)

func TestSanitizer_Sanitize(t *testing.T) {
	s := NewSanitizer()

	html := `
		<html>
			<head><title>Test Page</title></head>
			<body>
				<nav>Menu Items</nav>
				<article>
					<h1>Article Title</h1>
					<p>This is a paragraph with <a href="#">a link</a>.</p>
					<script>alert('bad');</script>
					<style>body { color: red; }</style>
				</article>
				<footer>Footer Content</footer>
			</body>
		</html>
	`

	clean, err := s.Sanitize(html)
	if err != nil {
		t.Fatalf("Sanitize failed: %v", err)
	}

	// Should not contain scripts or styles
	if strings.Contains(clean, "alert('bad')") {
		t.Error("Cleaned text still contains script content")
	}
	if strings.Contains(clean, "color: red") {
		t.Error("Cleaned text still contains style content")
	}

	// Should not contain nav or footer
	if strings.Contains(clean, "Menu Items") {
		t.Error("Cleaned text still contains nav content")
	}
	if strings.Contains(clean, "Footer Content") {
		t.Error("Cleaned text still contains footer content")
	}

	// Should contain the main article content
	if !strings.Contains(clean, "Article Title") {
		t.Error("Cleaned text missing article title")
	}
	if !strings.Contains(clean, "This is a paragraph") {
		t.Error("Cleaned text missing paragraph content")
	}

	// Check whitespace cleaning (should be joined with single newlines)
	if strings.Contains(clean, "\n\n") {
		t.Error("Cleaned text contains redundant newlines")
	}
}
