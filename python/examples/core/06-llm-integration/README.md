# 06 — llm-integration

A paired example that bridges **Canvus notes** with a **local Ollama LLM**.
Two entry points share the same package:

- `watcher.py` — subscribes to `CANVUS_CANVAS_ID`. When a note appears whose
  text starts with `?`, it strips the leading `?`, asks Ollama, and posts the
  answer as a sibling note 400 px to the right (background `#1d71b8ff`).
- `responder.py` — one-shot diagnostic. Polls Ollama's `/api/tags` and prints
  the installed models. Use it to verify Ollama before starting the watcher.

## Prerequisites

- A running Ollama instance (default `http://localhost:11434`).
- A model installed (default `llama3.2`). Install with:
  ```bash
  ollama pull llama3.2
  ```
- The usual Canvus creds.

## Environment

| Env var            | Required | Purpose                              |
| ------------------ | -------- | ------------------------------------ |
| `CANVUS_API_URL`   | yes      | Canvus base URL.                     |
| `CANVUS_API_KEY`   | yes      | API token with edit access.          |
| `CANVUS_CANVAS_ID` | yes      | Canvas to watch + answer on.         |
| `OLLAMA_URL`       | no       | Default `http://localhost:11434`.    |
| `OLLAMA_MODEL`     | no       | Default `llama3.2`.                  |

## Run

Verify Ollama first:

```bash
python examples/core/06-llm-integration/responder.py
```

Then start the watcher (long-lived):

```bash
python examples/core/06-llm-integration/watcher.py
```

In another terminal (or in the Canvus UI), create a note whose text starts
with `?`, e.g. `?What is the capital of Hungary?`. Within a few seconds an
answer note appears next to it.

## Architectural choices

- **Deduplication via in-memory set.** The first batch the subscribe
  endpoint returns is a *snapshot* of every existing note; treating those as
  brand-new "questions" would re-answer everything on each restart. The
  watcher tracks every note id it has seen and only answers each one once.
  This is intentionally process-local; restarting the watcher will re-answer
  any pre-existing question note. For a durable bridge, persist `seen` to
  disk or use a server-side filter.
- **Single shared `httpx.AsyncClient`.** The watcher uses one client for
  the lifetime of the process to benefit from HTTP keep-alive against
  Ollama.
- **Polite truncation.** Answer notes are clipped to 4000 chars to avoid
  posting a wall of text into the canvas.

## Troubleshooting

- **Watcher prints `ollama error: …`:** run `responder.py` first; if it also
  fails, Ollama is not reachable at `OLLAMA_URL`.
- **Answer never appears:** the API key may lack edit access on the
  canvas — check with example 03.
- **Wrong model name:** Ollama returns 404 for unknown models. Run
  `responder.py` to see which models are installed.
