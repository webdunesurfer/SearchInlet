# ADR 001: Embedded SQLite Database

## Status
Accepted

## Context
SearchInlet has pivoted from a multi-tenant SaaS to a self-hosted open-source utility. We need a way to store access tokens and simple usage logs (e.g., rate limits). Managing an external database (like PostgreSQL) adds unnecessary friction to the installation process.

## Decision
We will use **SQLite** (via GORM) as the sole database for the project. 

## Consequences
*   **Positive:** Zero configuration. The database is a single local file, making deployment trivial.
*   **Positive:** Drastically reduced memory and infrastructure footprint.
*   **Negative:** Scaling horizontally across multiple servers is not natively supported, which aligns with our decision to focus on single-node self-hosting.
