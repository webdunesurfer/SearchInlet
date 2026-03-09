# Phase 1: Core Foundation

The goal of Phase 1 is to build a functional MCP-to-SearXNG gateway that provides high-quality, LLM-optimized search results.

## 1.1 SearXNG Integration (`internal/searxng`)
- [x] Implement a robust JSON client for the SearXNG API.
- [x] Support for multiple SearXNG backend URLs (load balancing/failover).
- [x] Configuration for specific search engines (Google, Bing, etc.) via environment variables.

## 1.2 Optimization Pipeline (`internal/optimizer`)
- [x] **Sanitizer:** Remove HTML boilerplate, scripts, and CSS using `bluemonday`.
- [x] **Truncator:** Implement token-aware truncation using `tiktoken-go`.
- [ ] **Ranker:** Basic relevance scoring of snippets based on the search query.

## 1.3 MCP Server (`internal/mcp`)
- [x] Implement `search` tool with query and limit parameters.
- [ ] Implement `get_page_content` tool for deep-scraping specific URLs from search results.
- [ ] Add structured logging (slog) for debugging tool calls.

## 1.4 Infrastructure
- [x] Dockerfile for easy deployment of the SearchInlet gateway.
- [x] Basic CI/CD (GitHub Actions) for linting and building Go binaries.
- [x] One-command installation script for quick server deployment.

## 1.5 Verification & Testing
- [x] **Unit Tests:** 100% coverage for `internal/optimizer` (sanitization, truncation logic).
- [x] **Integration Test:** Scripted search against a real SearXNG instance using `go test`.
- [x] **Manual Verification:** Use `npx @modelcontextprotocol/inspector` on the local binary.
- [x] **VPS Deployment:** Deploy the Phase 1 Docker container to a remote VPS and verify connectivity via MCP-over-SSE (or SSH tunnel).
