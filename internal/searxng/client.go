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
	Unresponsive    []string `json:"unresponsive_engines,omitempty"`
}

// Client is a SearXNG API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new SearXNG client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// SearchOptions provides optional parameters for the search
type SearchOptions struct {
	Engines  []string
	Page     int
	Language string
}

// Search performs a search query against the SearXNG instance
func (c *Client) Search(ctx context.Context, query string, opts *SearchOptions) (*SearchResponse, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// SearXNG usually expects the JSON output at /search or just the root with format=json
	q := u.Query()
	q.Set("q", query)
	q.Set("format", "json")

	if opts != nil {
		if len(opts.Engines) > 0 {
			// SearXNG expects engines separated by comma or multiple engine parameters
			// Standard way is comma-separated string in "engines" param
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
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResp, nil
}
