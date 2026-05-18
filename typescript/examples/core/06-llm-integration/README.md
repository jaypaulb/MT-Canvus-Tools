# Example 06 — llm-integration

Paired apps that bridge Canvus and a local Ollama install. Ask a
question by creating a note that starts with `?`; the watcher posts
the LLM's answer as a sibling note.

## Purpose

Two independent entry points:

- **`src/watcher.ts`** — subscribes to notes on `CANVUS_CANVAS_ID`.
  When a *new* note appears whose text starts with `?`, sends the
  question to Ollama's `/api/generate` and creates an "answer" note
  400px to the right with a blue (`#1d71b8ff`) background.
- **`src/responder.ts`** — diagnostic only. Lists the Ollama models
  available via `/api/tags`. Run this first to confirm Ollama is
  reachable and that the model you want is pulled.

## Prerequisites

- Node 20+ and pnpm 9+
- A reachable Canvus server with an API key that has edit access to
  `CANVUS_CANVAS_ID`
- A local Ollama install (https://ollama.com), with at least one
  model pulled (`ollama pull llama3.2`)

## Configuration

| Variable           | Required | Default                    | Description                       |
| ------------------ | -------- | -------------------------- | --------------------------------- |
| `CANVUS_BASE_URL`  | yes      |                            | API base URL                      |
| `CANVUS_API_KEY`   | yes      |                            | API key with edit access          |
| `CANVUS_CANVAS_ID` | yes      |                            | Canvas to watch                   |
| `OLLAMA_URL`       | no       | `http://localhost:11434`   | Ollama API base                   |
| `OLLAMA_MODEL`     | no       | `llama3.2`                 | Model name to query               |

## Run

### Step 1 — confirm Ollama is reachable

```bash
pnpm dev:responder
# {"level":30,"component":"example-llm-responder","url":"http://localhost:11434/api/tags","msg":"querying Ollama tags"}
# {"level":30,"component":"example-llm-responder","modelCount":3,"msg":"available models"}
#   - llama3.2:latest
#   - mistral:latest
#   - ...
```

### Step 2 — start the watcher

```bash
pnpm dev:watcher
```

Leave it running. In the Canvus desktop client (or via example 03),
create a note whose text starts with `?`, e.g. `?What is 2 + 2?`.
Within a few seconds an answer note should appear immediately to the
right.

Press Ctrl-C to stop the watcher.

## How it works

The watcher uses the same async-iterator streaming pattern as
example 05, layered with:

- **Dedup by widget id** — the initial subscribe snapshot contains
  every existing note. An in-memory `Set` tracks widget ids already
  processed, so pre-existing questions are not re-answered on
  startup. Restarting the watcher will answer every `?`-note on the
  canvas again — this is intentional for a demo; persist the set in
  production.
- **Self-loop suppression** — when the watcher posts an answer note,
  it adds the new note's id to the dedup set immediately so the
  upcoming stream event for the answer does not trigger another LLM
  call.
- **Native `fetch`** — Node 20+ ships `fetch` natively; no extra
  dependency is needed to talk to Ollama.

Both entry points are declared in `package.json` under the `bin`
field, so a `pnpm install`-driven workspace can symlink them as
`example-llm-watcher` / `example-llm-responder` once built.

## Troubleshooting

- **"Ollama returned non-2xx"** — Ollama is not running, or
  `OLLAMA_URL` is wrong. Try `curl http://localhost:11434/api/tags`.
- **"model 'X' not found"** — pull it: `ollama pull llama3.2`.
- **No answer note appears** — check the watcher log: was the question
  detected? Did the Ollama call succeed? Did the note POST fail with
  a 403?
- **Watcher answers every existing `?`-note on restart** — by design
  (in-memory dedup). Use a persistent store (sqlite, Redis) if you
  need restart safety.
