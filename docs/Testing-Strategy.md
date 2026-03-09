# Testing Strategy

This document outlines the testing process for **SearchInlet** to ensure that each phase of development is stable and ready for production.

---

## 1. Testing Environments

### 1.1 Local Environment
*   **Unit Tests:** Every core logic component (especially `internal/optimizer` and `internal/auth`) must have high unit test coverage.
*   **MCP Inspector:** Use the official `npx @modelcontextprotocol/inspector` to manually verify tool calls.
*   **Local SearXNG:** A local Dockerized instance of SearXNG for rapid iteration.

### 1.2 Remote Testing (VPS: 194.163.160.234)
The VPS is used for "real-world" integration and multi-user testing.

*   **Deployment Method:** GitHub Actions push Docker images to the VPS or pull from a private registry.
*   **Integration Tests:** Go-based integration tests that query the remote VPS endpoints.
*   **Dashboard Verification:** Manual and automated testing of the user interface.
*   **SSE Testing:** Testing the **Server-Sent Events (SSE)** transport for MCP, ensuring that it works over a real network (handling latency, timeouts, and reconnections).

---

## 2. Testing Procedures per Phase

### Phase 1: Core Foundation
- **Goal:** Robust search translation and sanitization.
- **Key Test:** `go test ./internal/searxng/...` and `go test ./internal/optimizer/...`.
- **VPS Test:** Deploy Phase 1 as a single-binary MCP server and connect to it using an SSH tunnel or local port forwarding.

### Phase 2: User & Service Layer
- **Goal:** Multi-tenant security and management.
- **Key Test:** Mocked PostgreSQL/Redis tests for API key validation and rate-limiting.
- **VPS Test:** Deploy the dashboard and API gateway. Verify that unauthorized keys are rejected and that usage stats are correctly recorded in the database.

### Phase 3: Billing & Scale
- **Goal:** Monetization and high availability.
- **Key Test:** Automated billing cycle simulation (Stripe webhooks).
- **VPS Test:** Load testing using a tool like `k6` to simulate high concurrent usage, ensuring the system doesn't crash or leak memory under stress.

---

## 3. Continuous Integration (CI)
Every commit to the `main` branch must pass:
1.  `go fmt ./...`
2.  `go vet ./...`
3.  `go test ./...`
4.  `go build` (to ensure no compilation errors)
