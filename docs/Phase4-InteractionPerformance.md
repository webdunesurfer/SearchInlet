# Phase 4: Interaction & Performance

The goal of Phase 4 is to make SearchInlet a more interactive and capable research tool by allowing agents to read full web pages and by optimizing the distillation process through internal streaming and progress reporting.

## 4.1 Deep Web Reading (`read_page`)
- [x] Implemented `internal/reader` package for clean web scraping.
- [x] Added `read_page(url string)` tool to MCP server.
- [x] Integrated with distillation pipeline to summarize long articles.
- [x] Automatic boilerplate removal (nav, footer, ads).

## 4.2 Streaming & Responsiveness
- [x] Updated `OllamaClient` to support `stream: true` generating real-time chunks.
- [x] Implemented internal streaming in search handlers to prevent "frozen" backend state.
- [x] Periodic logging of distillation progress to server console.
- [ ] Implement `notifications/progress` support for compatible MCP clients.

## 4.3 UI & Metrics
- [x] Added Performance Overview to dashboard.
- [x] Support for custom model selection and removal.
- [x] Real-time download progress bars for Ollama models.
