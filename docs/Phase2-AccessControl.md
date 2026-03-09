# Phase 2: Access Control & Admin Layer

The goal of Phase 2 is to add a simple, lightweight security and token management layer to the self-hosted utility. This ensures that while the MCP server is accessible over the network (via SSE), only authorized agents can use it.

## 2.1 Embedded Database (`internal/db`)
- [x] Implement a simple **SQLite** database via GORM. No external database dependencies.
- [x] Define `Token`, `UsageLog`, and `LoginAttempt` schemas.

## 2.2 Token Management & Rate Limiting (`internal/auth`)
- [x] Implement a mechanism to generate secure Access Tokens (e.g., one for Cursor, one for a local script).
- [x] Implement a lightweight SQLite-backed rate limiter (e.g., configurable daily limit per token).
- [x] Add HTTP Middleware to validate the token in the `Authorization` header before opening an SSE connection or processing POST messages.
- [x] **Brute Force Protection:** Track failed admin login attempts and temporarily ban IPs after 5 failures.

## 2.3 Admin Dashboard (`internal/dashboard`)
- [x] Build a modern, Server-Side Rendered (SSR) dashboard using Go Templates.
- [x] Random secure Admin password generated during installation.
- [x] Interface to manage tokens, view generated strings once, and track basic usage stats per token.

## 2.4 MCP Server Transport (`internal/mcp`)
- [x] Implement **Server-Sent Events (SSE) transport** using the official Go MCP SDK handler.
- [x] Integrate **Caddy** as a reverse proxy to provide automatic HTTPS/SSL for the dashboard and SSE endpoint.
