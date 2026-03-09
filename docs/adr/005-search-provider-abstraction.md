# ADR 005: Search Backend Provider Abstraction and Scaling Strategy

## Status
Deprecated (Replaced by single-tenant utility focus; proxy pools not required)

## Context
SearchInlet currently relies on a single, locally hosted SearXNG instance. While excellent for privacy and initial development, relying on a single IP address to scrape major search engines (Google, DuckDuckGo) is not viable for a multi-tenant SaaS. As query volume increases, search engines will inevitably block the server's IP address with rate limits (`429 Too Many Requests`) or CAPTCHAs.

We need a strategy that allows SearchInlet to scale indefinitely while maintaining a predictable cost structure that we can pass on to users via API token pricing.

## Decision
1.  **Search Provider Abstraction Layer:** We will refactor the codebase to introduce a generic `SearchProvider` interface. While SearXNG is our primary engine, this abstraction ensures the system is decoupled from any single backend. This allows us to inject alternative engines (e.g., Brave Search API, Tavily, or "Bring Your Own Key" options) in the future if needed.
2.  **Primary Scaling Mechanism (Proxy Pools):** To scale our core service, we will implement **Rotating Proxy Pools** within our SearXNG infrastructure. By routing outbound SearXNG requests through residential or datacenter proxy pools (e.g., Smartproxy, Oxylabs), each search will appear to originate from a different global IP address.
3.  **Service Economy:** The cost of the service (what we charge users per token/query) will be directly calculated based on the bandwidth and request costs of these proxy pools, plus the cost of the optional Phase 3 Distillation compute.

## Consequences
*   **Positive (Scalability):** We can serve thousands of concurrent agents without being IP-banned by Google or Bing.
*   **Positive (Architecture):** The Go codebase remains clean. The abstraction layer means the business logic (routing, billing, distillation) doesn't care *how* the search was performed.
*   **Positive (Business Model):** We have a clear path to profitability by understanding the exact Unit Economic cost of a search (Proxy Cost + Compute Cost).
*   **Negative (Infrastructure Complexity):** We must manage proxy authentication, handle proxy failures, and potentially deploy multiple SearXNG instances across different regions to balance the load.
