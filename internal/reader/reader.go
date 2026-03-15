package reader

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
)

type Reader struct {
	client    *http.Client
	sanitizer *bluemonday.Policy
}

func NewReader() *Reader {
	return &Reader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		sanitizer: bluemonday.StrictPolicy(),
	}
}

func (r *Reader) ReadURL(ctx context.Context, url string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", err
	}

	// Use a common browser user-agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("received status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	title := doc.Find("title").Text()
	
	// Remove known noise tags
	doc.Find("script, style, nav, footer, header, noscript, iframe, svg, aside, .ads, .menu, .navigation").Remove()

	// Extract text from the "best" containers first
	var mainContent strings.Builder
	containers := []string{"article", "main", "#content", ".post-content", ".article-content", "body"}
	
	found := false
	for _, selector := range containers {
		if node := doc.Find(selector); node.Length() > 0 {
			mainContent.WriteString(r.extractText(node))
			found = true
			break
		}
	}

	if !found {
		mainContent.WriteString(r.extractText(doc.Find("body")))
	}

	return title, strings.TrimSpace(mainContent.String()), nil
}

func (r *Reader) extractText(s *goquery.Selection) string {
	var sb strings.Builder
	s.Find("h1, h2, h3, h4, h5, h6, p, li").Each(func(i int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		if text == "" {
			return
		}

		// Add formatting markers
		tag := goquery.NodeName(sel)
		if strings.HasPrefix(tag, "h") {
			sb.WriteString("\n\n# " + text + "\n")
		} else if tag == "li" {
			sb.WriteString("\n* " + text)
		} else {
			sb.WriteString("\n\n" + text)
		}
	})
	return sb.String()
}
