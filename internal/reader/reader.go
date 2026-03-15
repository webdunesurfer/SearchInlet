package reader

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Reader struct {
	client *http.Client
}

func NewReader() *Reader {
	return &Reader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

var blockedRanges = []*net.IPNet{
	parseCIDR("127.0.0.0/8"),    // Loopback
	parseCIDR("10.0.0.0/8"),     // Private-A
	parseCIDR("172.16.0.0/12"),  // Private-B
	parseCIDR("192.168.0.0/16"), // Private-C
	parseCIDR("169.254.0.0/16"), // Link-local
	parseCIDR("::1/128"),        // IPv6 Loopback
	parseCIDR("fe80::/10"),      // IPv6 Link-local
	parseCIDR("fc00::/7"),       // IPv6 Unique local
}

func parseCIDR(s string) *net.IPNet {
	_, n, _ := net.ParseCIDR(s)
	return n
}

func isBlocked(ip net.IP) bool {
	for _, block := range blockedRanges {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func (r *Reader) ReadURL(ctx context.Context, targetURL string) (string, string, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %w", err)
	}

	// Resolve IP to prevent SSRF
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve hostname: %w", err)
	}

	for _, ip := range ips {
		if isBlocked(ip) {
			return "", "", fmt.Errorf("access to private network address %s is blocked", ip.String())
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
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
	s.Find("h1, h2, h3, h4, h5, h6, p, li, table").Each(func(i int, sel *goquery.Selection) {
		tag := goquery.NodeName(sel)
		
		if tag == "table" {
			sb.WriteString("\n\n[TABLE]\n")
			sel.Find("tr").Each(func(j int, tr *goquery.Selection) {
				var row []string
				tr.Find("th, td").Each(func(k int, cell *goquery.Selection) {
					row = append(row, strings.TrimSpace(cell.Text()))
				})
				if len(row) > 0 {
					sb.WriteString("| " + strings.Join(row, " | ") + " |\n")
				}
			})
			return
		}

		text := strings.TrimSpace(sel.Text())
		if text == "" {
			return
		}

		// Add formatting markers
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
