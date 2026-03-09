package searxng

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Result represents a single search result from SearXNG
type Result struct {
	Title           string   `json:"title"`
	URL             string   `json:"url"`
	Content         string   `json:"content"`
	Engine          string   `json:"engine"`
	Score           float64  `json:"score"`
	Engines         []string `json:"engines"`
	Category        string   `json:"category"`
	PublishedDate   string   `json:"publishedDate,omitempty"`
	Thumbnail       string   `json:"thumbnail,omitempty"`
}

// SearchResponse represents the full JSON response from SearXNG
type SearchResponse struct {
	Query           string   `json:"query"`
	NumberOfResults int      `json:"number_of_results"`
	Results         []Result `json:"results"`
	Suggestions     []string `json:"suggestions"`
	Unresponsive    []any    `json:"unresponsive_engines,omitempty"`
}

// Client is a SearXNG API client
type Client struct {
	BaseURLs   []string
	HTTPClient *http.Client
}

// NewClient creates a new SearXNG client with support for multiple URLs
func NewClient(baseURLs []string) *Client {
	return &Client{
		BaseURLs:   baseURLs,
		HTTPClient: &http.Client{},
	}
}

// SearchOptions provides optional parameters for the search
type SearchOptions struct {
	Engines  []string
	Page     int
	Language string
}

// Search performs a search query against the available SearXNG instances with failover
func (c *Client) Search(ctx context.Context, query string, opts *SearchOptions) (*SearchResponse, error) {
	var lastErr error

	for _, baseURL := range c.BaseURLs {
		u, err := url.Parse(baseURL)
		if err != nil {
			lastErr = fmt.Errorf("invalid base URL %s: %w", baseURL, err)
			continue
		}

		q := u.Query()
		q.Set("q", query)
		q.Set("format", "json")

		if opts != nil {
			if len(opts.Engines) > 0 {
				enginesStr := ""
				for i, e := range opts.Engines {
					if i > 0 {
						enginesStr += ","
					}
					enginesStr += e
				}
				q.Set("engines", enginesStr)
			}
			if opts.Page > 0 {
				q.Set("pageno", strconv.Itoa(opts.Page))
			}
			if opts.Language != "" {
				q.Set("language", opts.Language)
			}
		}

		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request for %s: %w", baseURL, err)
			continue
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed for %s: %w", baseURL, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, baseURL)
			continue
		}

		var searchResp SearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			lastErr = fmt.Errorf("failed to decode response from %s: %w", baseURL, err)
			continue
		}

		// Success! Return the response
		return &searchResp, nil
	}

	return nil, fmt.Errorf("all SearXNG instances failed. last error: %w", lastErr)
}
