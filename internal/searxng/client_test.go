package searxng

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Search(t *testing.T) {
	// Create a mock SearXNG server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for mandatory search parameters
		if r.URL.Query().Get("q") == "" {
			t.Errorf("Expected 'q' parameter")
		}
		if r.URL.Query().Get("format") != "json" {
			t.Errorf("Expected 'format=json'")
		}

		// Mock response
		resp := SearchResponse{
			Query:           r.URL.Query().Get("q"),
			NumberOfResults: 1,
			Results: []Result{
				{
					Title:   "Search Result 1",
					URL:     "https://example.com/1",
					Content: "Snippet for result 1",
					Engine:  "google",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient([]string{server.URL})
	resp, err := client.Search(context.Background(), "test query", &SearchOptions{
		Engines: []string{"google", "bing"},
	})

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if resp.Query != "test query" {
		t.Errorf("Expected query 'test query', got '%s'", resp.Query)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(resp.Results))
	}

	if resp.Results[0].Title != "Search Result 1" {
		t.Errorf("Expected title 'Search Result 1', got '%s'", resp.Results[0].Title)
	}
}
