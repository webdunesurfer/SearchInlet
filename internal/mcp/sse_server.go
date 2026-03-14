package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/webdunesurfer/SearchInlet/internal/auth"
	"github.com/webdunesurfer/SearchInlet/internal/db"
	"github.com/webdunesurfer/SearchInlet/internal/distiller"
	"github.com/webdunesurfer/SearchInlet/internal/optimizer"
	"github.com/webdunesurfer/SearchInlet/internal/searxng"
	"gorm.io/gorm"
)

type SSEServer struct {
	mcpServer    *mcp.Server
	searxng      *searxng.Client
	sanitizer    *optimizer.Sanitizer
	truncator    *optimizer.Truncator
	tokenManager *auth.TokenManager
	distiller    *distiller.OllamaClient
	db           *gorm.DB
}

func NewSSEServer(name, version, searxngURL string, tm *auth.TokenManager, database *gorm.DB, ollamaURL string) (*SSEServer, error) {
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
		distiller:    distiller.NewOllamaClient(ollamaURL),
		db:           database,
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

func (s *SSEServer) getSetting(key, defaultValue string) string {
	var setting db.GlobalSetting
	if err := s.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return defaultValue
	}
	return setting.Value
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

	metrics := auth.UsageMetrics{}
	startTime := time.Now()

	resp, err := s.searxng.Search(ctx, args.Query, &searxng.SearchOptions{
		Engines: args.Engines,
	})
	metrics.SearchLatencyMS = time.Since(startTime).Milliseconds()

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

	rawText := strings.Join(snippets, "\n---\n")
	metrics.InputTokens = s.truncator.CountTokens(rawText)

	optimizedSnippets := s.truncator.TruncateResults(snippets, maxTokens)
	finalText := strings.Join(optimizedSnippets, "\n---\n")

	if len(optimizedSnippets) == 0 {
		finalText = "No results found for your query."
	} else {
		// --- Distillation Logic ---
		distEnabled := s.getSetting("distillation_enabled", "false") == "true"
		if distEnabled {
			metrics.DistillationEnabled = true
			log.Printf("Distilling results for query: %s", args.Query)
			model := s.getSetting("distillation_model", "qwen2.5:3b")
			prompt := s.getSetting("distillation_prompt", "Summarize and extract the most relevant information from the following search results. Be concise and maintain technical accuracy.")
			
			distStart := time.Now()
			distilled, err := s.distiller.Distill(ctx, model, prompt, finalText)
			metrics.DistillLatencyMS = time.Since(distStart).Milliseconds()

			if err != nil {
				log.Printf("Distillation failed: %v", err)
				// Fallback to original text if distillation fails
			} else {
				finalText = distilled
			}
		}
	}

	metrics.OutputTokens = s.truncator.CountTokens(finalText)

	// Record metrics in background
	if tokenID, ok := auth.GetTokenID(req.Context()); ok {
		go func() {
			_ = s.tokenManager.LogUsage(tokenID, "search", metrics)
		}()
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
