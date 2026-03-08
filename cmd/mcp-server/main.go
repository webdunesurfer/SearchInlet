package main

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchArgs defines the input for the search tool
type SearchArgs struct {
	Query string `json:"query" jsonschema:"description=The search query to perform via SearXNG"`
}

func main() {
	// 1. Initialize the server with implementation details
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "SearchInlet",
		Version: "0.1.0",
	}, nil)

	// 2. Register the search tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search",
		Description: "Search the internet via SearXNG",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
		// Validate query
		if args.Query == "" {
			return nil, nil, fmt.Errorf("query is required")
		}

		// TODO: Implement actual SearXNG search logic
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Searching for: %s (SearXNG client not yet implemented)", args.Query),
				},
			},
		}, nil, nil
	})

	log.Println("Starting SearchInlet MCP Server...")

	// 3. Run the server using the Stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
