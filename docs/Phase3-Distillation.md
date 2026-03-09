# Phase 3: Local Distillation

The goal of Phase 3 is to make the search context incredibly efficient and cheap for the end-user by utilizing a small, local LLM to summarize raw HTML data before passing it over the MCP protocol to the primary Agent.

## 3.1 Optimization Pipeline Upgrades (`internal/optimizer`)
- [ ] Implement smart text-chunking (splitting sanitized HTML into 1000-token blocks).
- [ ] Add regex/programmatic formatting enforcement for outputs from small, chatty models.

## 3.2 Ollama Integration (`internal/distillation`)
- [ ] Build a client to communicate with a local Ollama instance.
- [ ] Support models optimized for extraction (e.g., `qwen2.5:0.5b`, `llama3:8b`).
- [ ] Design few-shot prompts specifically tuned for sub-8B parameter models to perform "boolean filtering" (Does this text contain the answer? YES/NO) and "verbatim extraction".

## 3.3 Dashboard Enhancements
- [ ] Add a configuration page in the Admin Dashboard to specify the Ollama URL and the preferred Distillation model.
- [ ] Add an on/off toggle for the Distillation phase to allow users to bypass it if they prefer raw context.
