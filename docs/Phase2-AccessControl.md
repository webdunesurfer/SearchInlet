# Phase 2: Access Control & Admin Layer

The goal of Phase 2 is to add a simple, lightweight security and token management layer to the self-hosted utility. This ensures that while the MCP server is accessible over the network (via SSE), only authorized agents can use it.

## 2.1 Embedded Database (`internal/db`)
- [ ] Implement a simple **SQLite** database via GORM. No external database dependencies.
- [ ] Define `Token` and `UsageLog` schemas.

## 2.2 Token Management & Rate Limiting (`internal/auth`)
- [ ] Implement a mechanism to generate static Access Tokens (e.g., one for Cursor, one for a local script).
- [ ] Implement a lightweight, in-memory or SQLite-backed rate limiter (e.g., max 100 requests per day per token) to prevent accidental infinite loops from draining the local VPS.
- [ ] Add HTTP Middleware to validate the token in the `Authorization` header before opening an SSE connection.

## 2.3 Admin Dashboard (`internal/dashboard`)
- [ ] Build a minimal, Server-Side Rendered (SSR) dashboard using Go Templates and Tailwind CSS.
- [ ] Single Admin login (password generated on first boot or set via environment variable).
- [ ] Interface to view active tokens, generate new ones, and view basic usage stats (how many searches each token has performed).

## 2.4 MCP Server Transport (`internal/mcp`)
- [ ] Implement **Server-Sent Events (SSE) transport** over HTTP so external agents can connect without needing SSH access.
