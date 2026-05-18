# note-mapper

**note-mapper** is a web application that photographs a whiteboard or table covered in
physical Post-it notes and creates matching digital notes inside a Canvus canvas anchor zone,
preserving spatial layout from the photo.

A user selects a target canvas and anchor, takes a photo (or uploads one), and the app:

1. Preprocesses the image (resize, format-normalise).
2. Sends it to Google Gemini (multimodal LLM) which identifies each note, its text, colour,
   and pixel-space position.
3. Maps the pixel positions from image space into the anchor zone on the Canvus canvas.
4. Creates each note via the Canvus API.

## Quick start

```bash
# Required env vars
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-canvus-api-key
export GOOGLE_GENAI_API_KEY=your-gemini-key

# Build and run from the monorepo workspace root
cd go
go build -o bin/note-mapper ./examples/projects/note-mapper/cmd
./bin/note-mapper
```

Then open `http://localhost:8080` in a browser.

### Environment variables

| Variable | Description | Required |
|---|---|---|
| `CANVUS_API_URL` | Canvus API base URL (e.g. `https://host/api/v1`) | Yes (or via UI) |
| `CANVUS_API_KEY` | Canvus API key (`Private-Token` header value) | Yes (or via UI) |
| `GOOGLE_GENAI_API_KEY` | Google Gemini API key | Yes |
| `PORT` | HTTP listen port (default: `8080`) | No |
| `LOG_FORMAT` | `json` for structured logs; text otherwise | No |
| `LOG_LEVEL` | `debug`, `warn`, `error`, `info` (default `info`) | No |

Credentials can also be entered in the web UI's **Config** tab, which calls
`POST /api/set-credentials` to update them at runtime without restarting.

## API routes

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/upload-image` | Upload photo → LLM extraction → return notes |
| `POST` | `/api/scan-notes` | Re-scan last image with new zone params |
| `POST` | `/api/create-notes` | Place notes into a Canvus anchor zone |
| `POST` | `/api/set-credentials` | Update Canvus URL + API key in-session |
| `GET` | `/api/get-canvases` | List accessible Canvus canvases |
| `GET` | `/api/get-anchors?canvasID=` | List anchors for a canvas |
| `GET` | `/api/get-anchor-info?canvasID=&anchorID=` | Get anchor geometry |

## Architecture

```
cmd/main.go                 — Entry point; HTTP server setup; slog configuration
internal/
  api/handlers.go           — HTTP handlers for all /api/ routes
  canvus/client.go          — New SDK wrapper (replaces legacy canvusapi + mcs packages)
  config/config.go          — Config from env + in-session override
  image/processor.go        — Image preprocessing (resize, JPEG quality reduction)
  llm/
    extract.go              — Google Gemini multimodal extraction
    types.go                — Input/output types for LLM layer
  mapping/mapping.go        — Spatial mapping from image pixels → Canvus canvas coords
web/                        — Static HTML/CSS/JS front-end (unchanged from source)
```

### Coordinate system

**Canvus API coordinates are absolute pixels** relative to the canvas origin — not
normalised 0–1 values. Widget `location.x/y` and `size.width/height` are all pixel
values. The mapping layer handles the conversion from image pixel space to anchor-zone
canvas pixel space.

Anchors live at a fixed position and size in canvas pixels. Notes created inside an
anchor become children of that anchor (they have `parent_id = anchorID`) and their
coordinates are relative to the anchor's own top-left corner.

## Differences from the source repo (CanvusNoteMapper)

| Aspect | Source | This port |
|---|---|---|
| SDK | Hand-rolled `internal/canvusapi` (~490 LOC) + `internal/mcs` wrapper | New MT-Canvus-Tools Go SDK (`canvus.Session`) |
| Canvus API calls | Mix of SDK wrapper calls and raw `http.NewRequest` string-concat URLs in `mcs.go` | All via `session.ListCanvases`, `session.ListAnchors`, `session.GetAnchor`, `session.CreateNote` |
| Logging | `log.Printf` throughout | `log/slog` (structured, level-controlled, stderr) |
| Pipeline package | Incomplete 5-stage `internal/pipeline` referencing non-existent stages 2-5 (dead code) | Dropped entirely; direct `llm.ExtractPostitNotes` call is the real path |
| `internal/types` package | Duplicate of `llm` types | Dropped; `llm.Note` is the canonical in-process note type |
| Config | Global mutable package vars without env-var loading | Explicit `Load()` from env; `Set()` / `Get()` with mutex |
| Binary | `notescanner-linux-amd64` committed at repo root | Not ported |
| TS prototype | `ai/` directory (3 TypeScript files) | Dropped per audit recommendation |
| Gemini SDK | `github.com/google/generative-ai-go` (deprecated) | Same (still `generative-ai-go` — see SDK gaps below) |

## SDK gaps (Phase 4d candidates)

- **Gemini SDK migration**: source used `github.com/google/generative-ai-go/genai` (deprecated
  in favour of `github.com/google/genai`). This port uses the same deprecated package to
  keep parity — migration to `google/genai` is a Phase 4d item shared with ai-personas and
  translator (three sites, rule-of-three satisfied for extracting a shared helper).

- **`anchorData` in `app.js`**: the JS frontend populates anchor dimensions from
  `fetchAnchorInfo` calls but stores them in a module-scoped variable (`anchorData`) that
  is only populated inside `fetchAnchors`. This is a pre-existing UX limitation in the
  frontend — no server change needed.

## Limitations

- **Single-tenant**: the last uploaded image is stored in process memory; concurrent
  users will conflict.
- **Image dimensions hardcoded**: the LLM-extracted pixel coordinates assume a 1280×720
  reference frame. If the source image differs, the layout may be proportionally
  offset.
- **No auth on `/api/set-credentials`**: any client on the same network can change the
  server credentials. This is acceptable for a demo tool; add middleware if needed.
