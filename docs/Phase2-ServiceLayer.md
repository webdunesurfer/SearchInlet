# Phase 2: User & Service Layer

The goal of Phase 2 is to move from a standalone gateway to a SaaS-ready service with authentication, rate-limiting, and user management.

## 2.1 Database & Persistence (`internal/db`)
- [ ] Schema for Users, API Keys, and Usage Statistics (PostgreSQL recommended).
- [ ] Redis for session storage and high-speed rate-limiting.

## 2.2 Authentication & API Management (`internal/auth`)
- [ ] Implement multi-tenant API Key authentication for the MCP gateway.
- [ ] API Key creation, rotation, and revocation logic.
- [ ] **Rate-Limiting:** Per-user and per-key rate-limiting via Redis (Leaky Bucket or Sliding Window).

## 2.3 User Dashboard (`internal/dashboard`)
- [ ] Web-based UI (built with Go templates or a React/Vue SPA).
- [ ] User login/registration (support for email or GitHub OAuth).
- [ ] Interface to manage API keys and view usage history.

## 2.4 Usage Statistics (`internal/stats`)
- [ ] Track query volume, token usage (input/output), and latency.
- [ ] Export stats for dashboard visualization.
- [ ] Aggregation logic for daily/monthly usage reports.
