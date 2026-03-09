package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	searxngClient := searxng.NewClient(searxngURL)
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
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results (default 10)",
					"minimum":     1,
					"maximum":     50,
				},
				"engines": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "Specific search engines to use",
				},
				"max_tokens": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum tokens for the combined results (default 2000)",
				},
			},
			"required": []string{"query"},
		},
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

func (s *SSEServer) HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher.Flush()

	encoder := json.NewEncoder(w)

	message := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "init",
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"clientInfo": map[string]interface{}{
				"name":    "SearchInlet",
				"version": "1.0.0",
			},
		},
	}

	if err := encoder.Encode(message); err != nil {
		log.Printf("Error encoding message: %v", err)
		return
	}

	flusher.Flush()
}

func (s *SSEServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *SSEServer) RunHTTP(addr string) error {
	mux := http.NewServeMux()

	authMiddleware := auth.Middleware(s.tokenManager)

	mux.HandleFunc("/health", s.HandleHealth)
	mux.Handle("/sse", authMiddleware(http.HandlerFunc(s.HandleSSE)))

	log.Printf("Starting SSE server on %s", addr)
	return http.ListenAndServe(addr, mux)
}
