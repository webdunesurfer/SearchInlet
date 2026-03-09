package main

import (
	"context"
	"log"
	"os"

	"github.com/webdunesurfer/SearchInlet/internal/mcp"
)

func main() {
	// Configure SearXNG URL from environment variable or use a default
	searxngURL := os.Getenv("SEARXNG_URL")
	if searxngURL == "" {
		// Default to a public instance for testing, but recommend self-hosted
		searxngURL = "https://searxng.be/search"
		log.Printf("SEARXNG_URL not set, defaulting to %s", searxngURL)
	}

	// Initialize and run the MCP server
	server, err := mcp.NewServer("SearchInlet", "0.1.0", searxngURL)
	if err != nil {
		log.Fatalf("Failed to initialize MCP server: %v", err)
	}

	if err := server.Run(context.Background()); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
