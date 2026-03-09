# SearchInlet 🌊

**SearchInlet** is a high-performance, SaaS-ready **MCP (Model Context Protocol)** Gateway for **SearXNG**. It provides AI Agents with a secure, LLM-optimized interface for searching the internet, featuring advanced distillation, token management, and multi-tenant authentication.

---

## 🚀 Key Features

*   **MCP Native:** Built using the official Go MCP SDK for seamless integration with Claude Desktop, Cursor, and other AI Agents.
*   **SearXNG Powered:** Aggregates results from 70+ search engines while maintaining privacy.
*   **LLM Optimized:** 
    *   **Sanitization:** Strips boilerplate, HTML, and scripts for clean context.
    *   **Truncation:** Token-aware trimming using `tiktoken` to fit your model's context window.
    *   **Distillation:** Intelligent relevance scoring and snippet ranking.
*   **SaaS Ready:** (WIP) Built-in support for API Keys, Rate Limiting, and Usage Statistics.
*   **High Performance:** Written in Go for low-latency concurrent processing.

---

## 🛠 Architecture

SearchInlet acts as a bridge between your AI Agent and search backends.

```mermaid
graph LR
    Agent[AI Agent / MCP Client] <--> SI[SearchInlet Gateway]
    SI <--> SX[SearXNG Backend]
    
    subgraph "SearchInlet Internal"
        SI_MCP[MCP Server Layer]
        SI_Auth[Auth & Rate Limiting]
        SI_Proc[Optimization Pipeline]
        SI_DB[(Database)]
        SI_LLM((Local LLM<br>Ollama/Qwen))
    end
```

Detailed architecture can be found in [docs/Architecture.md](docs/Architecture.md).

---

## 📅 Roadmap & Phases

The project is being developed in three primary phases:

1.  **[Phase 1: Core Foundation](docs/Phase1-Core.md)** - Basic MCP gateway, SearXNG client, and sanitization logic.
2.  **[Phase 2: User & Service Layer](docs/Phase2-ServiceLayer.md)** - Auth, Rate-limiting, DB persistence, and User Dashboard.
3.  **[Phase 3: Billing & Scale](docs/Phase3-BillingScale.md)** - Stripe integration, advanced distillation (Ollama), and global scaling.

---

## 🏃 Quick Start (Linux / Server Deployment)

### Prerequisites
*   Docker & Docker Compose installed.

### Installation
Deploying the SearXNG backend and building the MCP server takes just one command.

```bash
curl -sSL https://raw.githubusercontent.com/webdunesurfer/SearchInlet/main/install.sh | bash
```

The script will automatically:
1. Clone the repository (if not already downloaded).
2. Generate a secure, random `SEARXNG_SECRET` to protect sessions.
3. Boot a local, privacy-respecting SearXNG instance using Docker.
4. Cross-compile the Go MCP server binary (no local Go installation required).

### Connecting your AI Agent (Cursor / Claude Desktop)
Since this is typically deployed on a remote server, you connect to it via an SSH tunnel in your Agent's MCP settings:

```json
{
  "mcpServers": {
    "searchinlet": {
      "command": "ssh",
      "args": [
        "user@your-server-ip",
        "SEARXNG_URL=http://localhost:8080/search /path/to/SearchInlet/bin/mcp-server-linux"
      ]
    }
  }
}
```

---

## 🧪 Testing with MCP Inspector

You can test the server using the official MCP Inspector:
```bash
npx @modelcontextprotocol/inspector ./bin/mcp-server
```

---

## 📄 License
This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**.
See the [LICENSE](LICENSE) file for details.
