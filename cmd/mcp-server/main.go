package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/webdunesurfer/SearchInlet/internal/auth"
	"github.com/webdunesurfer/SearchInlet/internal/dashboard"
	"github.com/webdunesurfer/SearchInlet/internal/db"
	"github.com/webdunesurfer/SearchInlet/internal/mcp"
)

func main() {
	searxngURL := os.Getenv("SEARXNG_URL")
	if searxngURL == "" {
		searxngURL = "https://searxng.be/search"
		log.Printf("SEARXNG_URL not set, defaulting to %s", searxngURL)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./searchinlet.db"
	}

	limitPerDay, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_PER_DAY"))
	if limitPerDay <= 0 {
		limitPerDay = 100
	}

	database, err := db.OpenDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	tokenManager := auth.NewTokenManager(database, limitPerDay)

	transportMode := os.Getenv("TRANSPORT_MODE")
	if transportMode == "" {
		transportMode = "stdio"
	}

	if transportMode == "sse" || transportMode == "admin" {
		sseServer, err := mcp.NewSSEServer("SearchInlet", "1.0.0", searxngURL, tokenManager)
		if err != nil {
			log.Fatalf("Failed to create SSE server: %v", err)
		}

		adminUser := os.Getenv("ADMIN_USER")
		if adminUser == "" {
			adminUser = "admin"
		}

		adminPass := os.Getenv("ADMIN_PASSWORD")
		if adminPass == "" {
			adminPass = "admin123"
		}

		dash := dashboard.NewDashboard(database, tokenManager, adminUser, adminPass)

		httpPort := os.Getenv("HTTP_PORT")
		if httpPort == "" {
			httpPort = "8080"
		}

		log.Printf("Starting SearchInlet on port %s", httpPort)
		log.Printf("Admin dashboard: http://localhost:%s/", httpPort)
		log.Printf("MCP SSE endpoint: http://localhost:%s/sse", httpPort)
		log.Printf("Health check: http://localhost:%s/health", httpPort)

		mux := http.NewServeMux()

		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})

		mux.Handle("/sse", sseServer.Handler())
		mux.Handle("/sse/", sseServer.Handler())

		mux.HandleFunc("/", dash.HandleHome)
		mux.HandleFunc("/login", dash.HandleLogin)
		mux.HandleFunc("/create-token", dash.HandleCreateToken)
		mux.HandleFunc("/revoke-token", dash.HandleRevokeToken)

		srv := &http.Server{
			Addr:    ":" + httpPort,
			Handler: mux,
		}

		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	} else {
		server, err := mcp.NewServer("SearchInlet", "1.0.0", searxngURL)
		if err != nil {
			log.Fatalf("Failed to initialize MCP server: %v", err)
		}

		if err := server.Run(context.Background()); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}
}
