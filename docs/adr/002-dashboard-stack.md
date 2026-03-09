# ADR 002: Dashboard Frontend Stack

## Status
Accepted

## Context
SearchInlet requires a web-based dashboard where users can register, log in, generate API keys, and view their usage statistics. We need to decide whether to build a separate Single Page Application (SPA) using a heavy framework like React/Vue, or keep the UI tightly coupled to the Go backend.

## Decision
We will build a Server-Side Rendered (SSR) dashboard utilizing **Go Templates**, enhanced with **HTMX** and **Alpine.js** for interactivity, and styled with **Tailwind CSS**.

## Consequences
*   **Positive:** **Single Binary Deployment.** The entire frontend can be compiled directly into the Go binary using the `embed` package. There is no need for a separate Node.js build step or serving static files via Nginx.
*   **Positive:** Rapid development velocity. We don't need to build REST/GraphQL APIs just to serve data to our own frontend; the Go handlers can render the data directly into the HTML.
*   **Positive:** HTMX and Alpine.js provide SPA-like interactivity (e.g., dynamic key generation, modal popups) without the overhead of a virtual DOM.
*   **Negative:** Developers must be comfortable writing raw HTML/CSS and using Go's templating syntax rather than component-based frameworks.
*   **Negative:** Less suitable if the dashboard becomes extremely complex (e.g., a full IDE or complex data visualization tool), though sufficient for typical SaaS account management.
