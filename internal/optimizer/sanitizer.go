package optimizer

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
)

// Sanitizer handles cleaning and stripping HTML content
type Sanitizer struct {
	policy *bluemonday.Policy
}

// NewSanitizer creates a new sanitizer with a strict policy
func NewSanitizer() *Sanitizer {
	// We want a very strict policy for LLMs - mostly plain text
	return &Sanitizer{
		policy: bluemonday.StrictPolicy(),
	}
}

// Sanitize strips all HTML tags and returns clean text
func (s *Sanitizer) Sanitize(html string) (string, error) {
	// 1. First, use goquery to remove non-content elements if this is a full page
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	// Remove scripts, styles, nav, footer, etc.
	doc.Find("script, style, nav, footer, iframe, noscript").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	// 2. Extract text with some structure preservation (optional, but good for LLMs)
	// For now, let's just get the text and clean it with bluemonday
	rawText, _ := doc.Html()
	cleanText := s.policy.Sanitize(rawText)

	// 3. Clean up whitespace
	return s.cleanWhitespace(cleanText), nil
}

// cleanWhitespace removes redundant newlines and spaces
func (s *Sanitizer) cleanWhitespace(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, "\n")
}
