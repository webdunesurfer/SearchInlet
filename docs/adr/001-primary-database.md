# ADR 001: Selection of Primary Database

## Status
Deprecated (Replaced by single-tenant SQLite strategy)

## Context
SearchInlet is evolving from a standalone local gateway into a multi-tenant SaaS platform. We need a reliable storage solution to persist critical business data, specifically:
*   User accounts and authentication details.
*   API Keys (hashed/encrypted) and their associated metadata.
*   Usage statistics and billing information (Phase 3).

The data model is highly relational (A User has many API Keys, an API Key has many Usage Records).

## Decision
We will use **PostgreSQL** as our primary relational database.

For local development and testing, we will utilize **SQLite** to reduce friction and eliminate the need to run a dedicated database container locally, leveraging our ORM (GORM) to abstract the differences between the two engines.

## Consequences
*   **Positive:** PostgreSQL is the industry standard for relational data, offering robust transaction support (ACID) and excellent JSON handling. It will easily support our Phase 3 horizontal scaling goals.
*   **Positive:** By using SQLite for local dev via GORM, developers can spin up the project instantly without managing external dependencies.
*   **Negative:** Requires managing database migrations and schema changes explicitly.
*   **Negative:** Adds a piece of infrastructure to manage in production deployments.
