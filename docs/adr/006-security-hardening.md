# ADR 006: Security Hardening (Brute Force Protection & HTTPS)

## Status
Accepted

## Context
As SearchInlet moves from a local SSH-tunneled tool to a service potentially accessible over the public internet, we must protect the Admin Dashboard from unauthorized access and brute-force attacks.

## Decision
1.  **Application-Level Brute Force Protection:** We have implemented a "Fail2Ban" style mechanism directly in the Go backend. Failed login attempts are recorded in SQLite by IP address. If an IP exceeds 5 failures within 15 minutes, it is temporarily banned from attempting further logins for 1 hour.
2.  **Reverse Proxy for SSL (Caddy):** Instead of building complex SSL/TLS certificate management into the Go binary, we recommend using **Caddy** as a reverse proxy. Caddy automatically handles Let's Encrypt certificates and provides a secure HTTPS layer for the application.
3.  **Secure Session Management:** Admin sessions use a cryptographically secure, random 32-byte token generated on server startup. Cookies are marked as `HttpOnly`, `SameSite=Strict`, and secure when served over HTTPS.

## Consequences
*   **Positive:** The Admin Dashboard is significantly more resistant to automated password guessing attacks.
*   **Positive:** Single-binary simplicity is maintained by offloading SSL certificates to a dedicated infrastructure component (Caddy).
*   **Negative:** Users must configure a reverse proxy or SSH tunnel to achieve a "Private" connection (HTTPS) if they want to avoid browser security warnings.
