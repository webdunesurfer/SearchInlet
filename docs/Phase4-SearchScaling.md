# Phase 4: Search Infrastructure Scaling

The goal of Phase 4 is to transition the search backend from a single-node, single-IP setup to a highly available, proxy-backed architecture that can handle thousands of concurrent users without being IP-banned by major search engines.

## 4.1 Search Provider Abstraction (`internal/search`)
- [ ] Refactor `internal/searxng` into a generic `internal/search` package.
- [ ] Create a `SearchProvider` Go interface with a standard `Search(query)` method.
- [ ] Implement the `SearXNGProvider` conforming to this interface.
- [ ] Implement a mock or fallback provider for testing.

## 4.2 Proxy Pool Integration
- [ ] Select a commercial proxy pool provider (e.g., Smartproxy, BrightData).
- [ ] Configure the SearXNG Docker image (`settings.yml` -> `outgoing.proxies`) to route requests through the proxy endpoints.
- [ ] Implement health checks within SearchInlet to monitor SearXNG proxy failures (e.g., detecting sudden spikes in empty results or 403s from the proxies themselves).

## 4.3 Unit Economics & Telemetry
- [ ] Implement telemetry to track the bandwidth (bytes in/out) used per search query via the proxies.
- [ ] Update the billing logic (from Phase 3) to calculate token pricing dynamically based on proxy usage costs.

## 4.4 High Availability SearXNG
- [ ] Deploy multiple SearXNG containers (or a swarm/Kubernetes cluster) behind a load balancer (e.g., Nginx, Traefik or HAProxy).
- [ ] Update SearchInlet to route requests to the load balancer rather than a single `localhost` instance.

## 4.5 Verification & Testing
- [ ] **Abstraction Test:** Ensure the generic `SearchProvider` interface works identically with different underlying implementations in unit tests.
- [ ] **Proxy IP Rotation Test:** Script a test that makes 5 sequential searches and verifies (via SearXNG debug logs or a test endpoint) that the outbound IP address changed for each request.
- [ ] **Load Test:** Sustain 5,000 queries per hour to ensure the proxy pool is effectively preventing `429 Too Many Requests` blocks from Google/DuckDuckGo.
