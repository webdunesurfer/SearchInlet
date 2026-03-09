# Phase 2: User & Service Layer

The goal of Phase 2 is to move from a standalone gateway to a SaaS-ready service with authentication, rate-limiting, and user management.

## 2.1 Database & Persistence (`internal/db`)
- [ ] Schema for Users, API Keys, and Usage Statistics using PostgreSQL.
- [ ] Implement database interactions using **GORM**.
- [ ] Redis for session storage and high-speed rate-limiting.

## 2.2 Authentication & API Management (`internal/auth`)
- [ ] Implement multi-tenant API Key authentication for the MCP gateway.
- [ ] API Key creation, rotation, and revocation logic.
- [ ] **Rate-Limiting:** Per-user and per-key rate-limiting via Redis (Leaky Bucket or Sliding Window).

## 2.3 User Dashboard (`internal/dashboard`)
- [ ] Web-based UI built with **Go Templates, HTMX, Alpine.js, and Tailwind CSS**.
- [ ] User login/registration (support for email or GitHub OAuth).
- [ ] Interface to manage API keys and view usage history.

## 2.4 MCP Server Transport (`internal/mcp`)
- [ ] Implement **Server-Sent Events (SSE) transport** over HTTPS.
- [ ] Middleware to extract and validate API keys from HTTP headers before establishing the SSE connection.

## 2.4 Usage Statistics (`internal/stats`)
- [ ] Track query volume, token usage (input/output), and latency.
- [ ] Export stats for dashboard visualization.
- [ ] Aggregation logic for daily/monthly usage reports.

## 2.5 Verification & Testing
- [ ] **API Auth Tests:** Verify that queries without valid API keys fail.
- [ ] **Rate-Limit Testing:** Automated script to flood the server and confirm 429 responses.
- [ ] **User Database Integrity:** Integration tests for key creation/revocation/usage tracking.
- [ ] **Dashboard Smoke Test:** Confirm user login and key management functionality on the VPS-hosted dashboard.
