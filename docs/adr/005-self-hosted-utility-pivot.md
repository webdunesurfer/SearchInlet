# ADR 005: Pivot to Self-Hosted Utility

## Status
Accepted

## Context
Initially, SearchInlet was planned as a multi-tenant SaaS with complex proxy pools, Stripe billing, and external PostgreSQL/Redis dependencies. However, providing a public search API at scale introduces significant risk of IP bans from upstream engines, and the operational overhead of a SaaS outweighs the immediate value proposition to the open-source community.

## Decision
We will simplify the architecture to be a **Self-Hosted Open-Source Utility**. 
* Access control will be limited to simple static tokens managed by a single admin.
* The application will run entirely on a single VPS alongside SearXNG.
* Complex scaling (proxy pools, load balancing) and commercial SaaS features are removed from the roadmap.

## Consequences
*   **Positive:** The project is drastically simpler to build, test, and maintain.
*   **Positive:** Users get maximum privacy and control over their data by running everything on their own hardware.
*   **Negative:** SearchInlet itself will not generate direct SaaS revenue.
