package searxng

import (
	"context"
	"os"
	"testing"
)

func TestClient_Integration(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	searxngURL := os.Getenv("SEARXNG_URL")
	if searxngURL == "" {
		t.Skip("Skipping integration test: SEARXNG_URL environment variable is not set")
	}

	client := NewClient([]string{searxngURL})

	// Perform a real search
	resp, err := client.Search(context.Background(), "Model Context Protocol", &SearchOptions{
		Engines: []string{"duckduckgo", "google"},
	})

	if err != nil {
		t.Fatalf("Integration search failed: %v", err)
	}

	// Validate the response structure
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}

	if resp.Query != "Model Context Protocol" {
		t.Errorf("Expected query 'Model Context Protocol', got '%s'", resp.Query)
	}

	if len(resp.Results) == 0 {
		t.Log("Warning: Received 0 results. This can happen with public instances due to rate limiting or engine issues, but the request was successful.")
	} else {
		// Check that at least the first result has basic fields populated
		firstResult := resp.Results[0]
		if firstResult.Title == "" {
			t.Error("Expected first result to have a title")
		}
		if firstResult.URL == "" {
			t.Error("Expected first result to have a URL")
		}
	}
}
