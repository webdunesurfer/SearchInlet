# Phase 1: Core Foundation

The goal of Phase 1 is to build a functional MCP-to-SearXNG gateway that provides high-quality, LLM-optimized search results.

## 1.1 SearXNG Integration (`internal/searxng`)
- [ ] Implement a robust JSON client for the SearXNG API.
- [ ] Support for multiple SearXNG backend URLs (load balancing/failover).
- [ ] Configuration for specific search engines (Google, Bing, etc.) via environment variables.

## 1.2 Optimization Pipeline (`internal/optimizer`)
- [ ] **Sanitizer:** Remove HTML boilerplate, scripts, and CSS using `bluemonday`.
- [ ] **Truncator:** Implement token-aware truncation using `tiktoken-go`.
- [ ] **Ranker:** Basic relevance scoring of snippets based on the search query.

## 1.3 MCP Server (`internal/mcp`)
- [ ] Implement `search` tool with query and limit parameters.
- [ ] Implement `get_page_content` tool for deep-scraping specific URLs from search results.
- [ ] Add structured logging (slog) for debugging tool calls.

## 1.4 Infrastructure
- [ ] Dockerfile for easy deployment of the SearchInlet gateway.
- [ ] Basic CI/CD (GitHub Actions) for linting and building Go binaries.

## 1.5 Verification & Testing
- [ ] **Unit Tests:** 100% coverage for `internal/optimizer` (sanitization, truncation logic).
- [ ] **Integration Test:** Scripted search against a real SearXNG instance using `go test`.
- [ ] **Manual Verification:** Use `npx @modelcontextprotocol/inspector` on the local binary.
- [ ] **VPS Deployment:** Deploy the Phase 1 Docker container to `194.163.160.234` and verify connectivity via MCP-over-SSE (or SSH tunnel).
