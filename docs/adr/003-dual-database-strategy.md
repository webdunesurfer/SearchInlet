# ADR 003: Dual Database Strategy (PostgreSQL + SQLite)

## Status
Accepted

## Context
SearchInlet needs a relational database to store Users, API Keys, and Usage Statistics. We evaluated options like PostgreSQL, SQLite, and in-memory stores. 

We face two conflicting requirements:
1.  **Production Readiness:** The production environment (VPS) requires high concurrency for simultaneous MCP API requests and robust transaction safety for billing.
2.  **Developer Experience:** Local development should be frictionless, ideally requiring "zero setup" so developers don't have to install and manage background database servers (like Dockerized Postgres) just to run tests or make small UI tweaks.

## Decision
We will adopt a **Dual Database Strategy** using **GORM**:

*   **Production Environment:** We will use **PostgreSQL**.
*   **Local/Test Environments:** We will use **SQLite**.

The application will dynamically select the database dialect based on the presence of the `DATABASE_URL` or `ENV=production` environment variables. Because we are using an ORM (GORM), the Go structs (schema) and query logic remain identical across both engines.

## Consequences
*   **Positive (Performance):** PostgreSQL effortlessly handles the concurrent read/write locks that occur when multiple agents update their usage statistics simultaneously.
*   **Positive (DX):** Developers can clone the repository, run `go run main.go`, and immediately have a working database (`searchinlet.db` file) without any infrastructure setup.
*   **Negative:** We must be careful not to use Postgres-specific features (like specific JSONB query operators or complex window functions) in our GORM queries, as they will cause the SQLite local environment to crash. All queries must rely on standard SQL supported by both dialects via GORM.
