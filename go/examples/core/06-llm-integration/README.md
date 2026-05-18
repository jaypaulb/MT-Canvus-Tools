# 06 — LLM Integration (Canvus + Ollama)

A pair of small binaries demonstrating an end-to-end LLM workflow on top of
a Canvus canvas:

- **`watcher`** — subscribes to a canvas's notes, watches for "question
  notes" (text starting with `?`), forwards each to a local Ollama LLM, and
  posts the answer back as a sibling note positioned to the right of the
  question.
- **`responder`** — a one-shot diagnostic that queries Ollama's `/api/tags`
  endpoint and prints installed models. Use this to verify Ollama is up
  before starting the watcher.

The workflow on the canvas, from a user's perspective, is dead simple:
- Type a note starting with `?` (e.g. `?What is 2+2?`).
- A green-blue MT-coloured answer note appears next to it within seconds.

## Prerequisites

- A local [Ollama](https://ollama.com) install with at least one model:
  ```bash
  ollama pull llama3.2
  ollama serve   # if not already running as a service
  ```
- API key with write access on the target canvas.

## Configuration

| Variable | Required | Description |
| --- | --- | --- |
| `CANVUS_BASE_URL` | yes | Full base URL including `/api/v1/`. |
| `CANVUS_API_KEY` | yes | API key with write access. |
| `CANVUS_CANVAS_ID` | yes (watcher) | UUID of canvas to watch. |
| `OLLAMA_URL` | no | Defaults to `http://localhost:11434`. |
| `OLLAMA_MODEL` | no | Defaults to `llama3.2`. |
| `LOG_FORMAT` | no | `text` (default) or `json`. |

## Running

### Step 1: verify Ollama
```bash
go run ./cmd/responder
```
Expected output:
```
level=INFO msg="ollama reachable" url=http://localhost:11434 model_count=2
llama3.2:latest                              2019393189 bytes  2026-05-01T...
```

If you see `connection refused`, Ollama is not running.
If `model_count=0`, run `ollama pull llama3.2`.

### Step 2: start the watcher
```bash
go run ./cmd/watcher
```
Expected output:
```
level=INFO msg="watcher starting" canvas_id=<id> ollama_url=http://localhost:11434 model=llama3.2
level=INFO msg="initial snapshot recorded" existing_note_count=4 existing_questions_skipped=0
```

### Step 3: ask a question on the canvas
In the Canvus client, create a new sticky note whose text starts with `?`:

> `?What's the capital of Hungary?`

Within a few seconds the watcher logs:
```
level=INFO msg="question received" question_id=<id> prompt="What's the capital of Hungary?"
level=INFO msg="ollama responded" question_id=<id> response_len=187
level=INFO msg="answer posted" question_id=<id> answer_id=<answer_id>
```
And a new MT-blue note appears 400px to the right of the question.

Ctrl-C stops the watcher cleanly.

## How it works

### Deduplication
The Canvus subscribe stream emits the full current note list on every
frame — these are snapshots, not deltas. Naively answering every question
in every frame would be a runaway loop. The watcher maintains an in-memory
`seen` map keyed by note ID:

1. The first frame is treated as the *initial snapshot*. Every note ID is
   recorded in `seen`, but no questions are answered. This means existing
   questions on the canvas at startup are ignored.
2. From frame 2 onwards, only IDs not in `seen` are eligible. New question
   notes (text starting with `?`) are forwarded to Ollama.
3. The answer note's own ID is added to `seen` immediately after creation,
   defensively — it shouldn't match the `?` prefix anyway, but a bug here
   could cause an infinite question/answer loop, so we make it impossible.

### Long-lived stream + per-request timeout
The session's `RequestTimeout` is set to 10 minutes (the subscribe
connection is long-lived). The Ollama POST is a separate `context.WithTimeout`
of 3 minutes — small models on CPU can take 30s+ for a single prompt.

### Why two binaries?
A single binary would tempt readers to launch Ollama queries before
proving connectivity. Splitting them encourages the run-responder-first
discipline that saves debugging time.

### Why `#1D71B8FF` for answers?
MT blue at full opacity — visually distinct from the canvas-default sticky
yellow so questions and answers are easy to scan visually.

## Architectural notes

- **No SDK Subscribe helper:** same pattern as example 05 — build the HTTP
  request directly using `session.HTTPClient`, which already has the API-key
  round-tripper installed.
- **Idempotency:** the watcher does not store state across restarts. If you
  Ctrl-C and restart, every existing note (including answered questions)
  becomes part of the new initial snapshot and is ignored.
- **No concurrent question handling:** questions are answered sequentially.
  An LLM call that takes 30s blocks the next question's start. For
  production use, dispatch each `answer` call into a goroutine and bound
  with a semaphore.

## Troubleshooting

| Symptom | Likely cause |
| --- | --- |
| `responder: connection refused` | Ollama is not running. Try `ollama serve`. |
| `responder: no models installed` | Run `ollama pull llama3.2` (or any model). |
| `watcher: 401` on subscribe | Wrong API key. |
| `watcher: 404` on subscribe | Wrong canvas ID. |
| Question added but no answer | Check the watcher log: is `question received` printed? If not, the note's text doesn't start with `?` (whitespace, BOM, etc). If yes, Ollama may be slow or stuck — check `ollama ps`. |
| Answer note appears but says nothing useful | Try a bigger model: `ollama pull llama3.1:8b` and set `OLLAMA_MODEL=llama3.1:8b`. |
