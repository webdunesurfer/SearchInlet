package mcp

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/webdunesurfer/SearchInlet/internal/optimizer"
	"github.com/webdunesurfer/SearchInlet/internal/searxng"
)

// Server wraps the MCP server and its dependencies
type Server struct {
	mcpServer *mcp.Server
	searxng   *searxng.Client
	sanitizer *optimizer.Sanitizer
	truncator *optimizer.Truncator
}

// SearchArgs defines the input for the search tool
type SearchArgs struct {
	Query     string   `json:"query" jsonschema:"description=The search query"`
	Limit     int      `json:"limit,omitempty" jsonschema:"description=Maximum number of results (default 10),minimum=1,maximum=50"`
	Engines   []string `json:"engines,omitempty" jsonschema:"description=Specific search engines to use"`
	MaxTokens int      `json:"max_tokens,omitempty" jsonschema:"description=Maximum tokens for the combined results (default 2000)"`
}

// NewServer initializes the MCP server with SearXNG and Optimizer components
func NewServer(name, version, searxngURL string) (*Server, error) {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Version: version,
	}, nil)

	searxngClient := searxng.NewClient(searxngURL)
	sanitizer := optimizer.NewSanitizer()
	truncator, err := optimizer.NewTruncator("gpt-4") // Default model
	if err != nil {
		return nil, fmt.Errorf("failed to initialize truncator: %w", err)
	}

	s := &Server{
		mcpServer: mcpServer,
		searxng:   searxngClient,
		sanitizer: sanitizer,
		truncator: truncator,
	}

	s.registerTools()

	return s, nil
}

func (s *Server) registerTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "search",
		Description: "Search the internet via SearXNG with LLM-optimized output",
	}, s.handleSearch)
}

func (s *Server) handleSearch(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
	if args.Query == "" {
		return nil, nil, fmt.Errorf("query is required")
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 10
	}

	maxTokens := args.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2000
	}

	// 1. Fetch results from SearXNG
	resp, err := s.searxng.Search(ctx, args.Query, &searxng.SearchOptions{
		Engines: args.Engines,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("searxng search failed: %w", err)
	}

	// 2. Process and optimize results
	var snippets []string
	for i, result := range resp.Results {
		if i >= limit {
			break
		}
		
		// Sanitize snippet content
		cleanContent, _ := s.sanitizer.Sanitize(result.Content)
		
		formatted := fmt.Sprintf("[%d] %s\nURL: %s\nSource: %s\nContent: %s\n", 
			i+1, result.Title, result.URL, result.Engine, cleanContent)
		snippets = append(snippets, formatted)
	}

	// 3. Truncate to token budget
	optimizedSnippets := s.truncator.TruncateResults(snippets, maxTokens)
	finalText := strings.Join(optimizedSnippets, "\n---\n")

	if len(optimizedSnippets) == 0 {
		finalText = "No results found for your query."
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: finalText,
			},
		},
	}, nil, nil
}

// Run starts the MCP server over Stdio
func (s *Server) Run(ctx context.Context) error {
	log.Printf("Starting SearchInlet MCP Server via Stdio...")
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}
