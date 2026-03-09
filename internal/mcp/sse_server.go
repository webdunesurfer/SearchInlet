package mcp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/webdunesurfer/SearchInlet/internal/auth"
	"github.com/webdunesurfer/SearchInlet/internal/optimizer"
	"github.com/webdunesurfer/SearchInlet/internal/searxng"
)

type SSEServer struct {
	mcpServer    *mcp.Server
	searxng      *searxng.Client
	sanitizer    *optimizer.Sanitizer
	truncator    *optimizer.Truncator
	tokenManager *auth.TokenManager
}

func NewSSEServer(name, version, searxngURL string, tm *auth.TokenManager) (*SSEServer, error) {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Version: version,
	}, nil)

	// Split URLs by comma
	urls := strings.Split(searxngURL, ",")
	for i := range urls {
		urls[i] = strings.TrimSpace(urls[i])
	}

	searxngClient := searxng.NewClient(urls)
	sanitizer := optimizer.NewSanitizer()
	truncator, err := optimizer.NewTruncator("gpt-4")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize truncator: %w", err)
	}

	s := &SSEServer{
		mcpServer:    mcpServer,
		searxng:      searxngClient,
		sanitizer:    sanitizer,
		truncator:    truncator,
		tokenManager: tm,
	}

	s.registerTools()
	return s, nil
}

func (s *SSEServer) registerTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "search",
		Description: "Search the internet via SearXNG with LLM-optimized output",
	}, s.handleSearch)
}

func (s *SSEServer) handleSearch(ctx context.Context, req *mcp.CallToolRequest, args SearchArgs) (*mcp.CallToolResult, any, error) {
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

	resp, err := s.searxng.Search(ctx, args.Query, &searxng.SearchOptions{
		Engines: args.Engines,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("searxng search failed: %w", err)
	}

	var snippets []string
	for i, result := range resp.Results {
		if i >= limit {
			break
		}

		cleanContent, _ := s.sanitizer.Sanitize(result.Content)

		formatted := fmt.Sprintf("[%d] %s\nURL: %s\nSource: %s\nContent: %s\n",
			i+1, result.Title, result.URL, result.Engine, cleanContent)
		snippets = append(snippets, formatted)
	}

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

func (s *SSEServer) Handler() http.Handler {
	// Create the official SDK SSE handler
	sseHandler := mcp.NewSSEHandler(func(req *http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)

	// Wrap it in our token authentication/rate-limiting middleware
	authMiddleware := auth.Middleware(s.tokenManager)
	
	return authMiddleware(sseHandler)
}
