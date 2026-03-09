# ADR 004: LLM Distillation Strategy and Model Selection

## Status
Accepted

## Context
In Phase 3, SearchInlet will implement a "Distillation" layer. This layer reads the noisy, raw HTML/text returned from SearXNG and synthesizes it into a dense, token-efficient summary before sending it over the MCP protocol to the user's primary AI Agent. 

We need to decide the approach for running this secondary distillation model, considering cost, speed, and hardware constraints (specifically, the desire to minimize or eliminate GPU costs by using ultra-small models like Qwen 0.5B/1.5B or Llama 1B/3B).

## Decision
We will build a flexible **Provider Abstraction Layer** in Go (`internal/distillation`) that supports querying both local models (via Ollama) and external APIs. 

For the **Local Model Strategy**, we will explicitly design our prompts and data-chunking pipeline to support **Sub-8B Parameter Models** (e.g., Qwen 1.5B, Llama-3-8B 4-bit).

Because ultra-small models struggle with long context windows and complex summarization, our distillation pipeline will use an **Extraction over Summarization** strategy:
1.  **Chunking:** The Go backend will split SearXNG results into small, manageable chunks (e.g., 500-1000 tokens).
2.  **Boolean Filtering / Extraction:** The small model will be prompted to perform simple tasks (e.g., "Extract verbatim quotes related to X" or "Reply YES if this text answers the prompt, otherwise NO").
3.  **Aggregation:** The Go backend will reassemble the extracted, high-signal chunks into the final MCP payload.

## Consequences
*   **Positive (Cost & Scalability):** By optimizing for sub-1B or sub-8B quantized models, SearchInlet can perform distillation entirely on CPU or cheap, low-VRAM GPUs. This keeps operating costs exceptionally low.
*   **Positive (Speed):** Tiny models process tokens in milliseconds, ensuring the MCP gateway remains responsive.
*   **Negative (Engineering Complexity):** We cannot rely on the "magic" of a GPT-4 class model to figure out formatting. The Go backend must handle aggressive chunking, few-shot prompt construction, and strict regex parsing to handle erratic outputs from small models.
