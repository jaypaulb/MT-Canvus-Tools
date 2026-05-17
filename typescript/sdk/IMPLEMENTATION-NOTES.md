# Implementation Notes — @mt-canvus-tools/sdk v0.1.0

These are the autonomous design decisions made during the initial
greenfield scaffold. Jaypaul was asleep; defaults below are reversible
in v0.2.

## 1. Validation strategy — trust the server (for now)

The SDK uses **zod for configuration parsing only** (env-var loading and
`createSession` options). It does **not** validate response payloads at
runtime.

Rationale:

- The API is the source of truth. Wrapping every response in a zod
  schema doubles the maintenance burden and silently rejects fields the
  server adds in a minor version.
- TypeScript types are wire-faithful (kebab-case keys, no transform);
  the cost of being wrong is a type assertion at the call site, not a
  silent corruption.
- The 150-endpoint surface area means duplicating each response schema
  would add ~3 000 lines of zod and ~10× the maintenance cost on every
  schema change.

When this should change: as soon as we add a response normaliser
(camelCase keys, Date parsing) we need runtime validation to make that
safe. v0.2 will reconsider.

## 2. Wire shape — preserved verbatim, no camelCase mapping

Field names in TypeScript types match the JSON exactly:
`canvas-id`, `widget-type`, `auto-text-color`. Indexed-access reads
(`canvas["canvas-id"]`) are mildly ugly but eliminate an entire class of
"the SDK silently dropped a field" bugs and make `curl` output directly
usable as a TypeScript value.

Reversible in v0.2 via a thin transform layer if Jaypaul prefers
camelCase.

## 3. Browser & Deno claims — declared but untested

The conventions doc says "intended to also work in browser and Deno
1.40+". This scaffold:

- Uses native `fetch` for ordinary requests (browser-safe).
- Uses `undici.request` for streaming (Node-only).
- Does not import any Node-specific module in the request/response path
  other than the streaming module.

**Tested:** Node 20 (vitest + the mock fetch in `tests/session.test.ts`).
**Not tested:** browser, Deno, Bun, Cloudflare Workers.

If browser support becomes a real requirement, `streamNdjson` needs a
fork that uses `ReadableStream` (the `response.body` reader) instead of
`undici`. The transport itself is already portable.

## 4. Streaming — undici + async generator

`streamNdjson<T>(transport, path, opts)` returns an `AsyncGenerator<T>`
that:

- Adds `subscribe=true` to the query string.
- Reads `undici`'s body iteratively (true backpressure, no full-body
  buffer).
- Splits on `\n`, trims, skips empty lines (the 15 s keepalive ping).
- Yields each JSON-parsed line.
- Tears down the connection on `break`/`return`/abort.

Open question: when the first frame is a JSON array (initial state for a
list endpoint), we yield the whole array as a single item. Some callers
may prefer per-item flattening. Deferred to v0.2 — easy to add a
`flatten: true` option.

## 5. Errors — discriminated union, exhaustive handling

The error hierarchy is:

```
CanvusError (kind: "api" | "validation" | "auth" | "network" | "not-found")
├── APIError (status, body)
│   └── NotFoundError (kind reassigned to "not-found")
├── AuthError (reason: "missing-token" | "expired" | "forbidden")
├── ValidationError (issues[])
└── NetworkError
```

`NotFoundError` is split out from generic `APIError` because 404 is
frequently normal control flow. The convention doc only specifies four
kinds; adding `not-found` is the only deviation from the locked
hierarchy. Convention amendments doc has not been edited (read-only per
boundary rules).

## 6. Source uncertainty — deferred for Phase 4 verification

- **Changelog §5 (RDP field naming):** the C++ serialiser may emit
  hyphenated keys (`host-id`, `content-id`, `connection-name`,
  `host-site`) but the public docs use underscored equivalents. This
  scaffold follows the public-docs convention. See the inline
  `RdpConnection` type comment.
  → Verify via `curl` against a live server before v0.2.

- **`UploadsFolderItem`:** the spec leaves the response shape "varies"
  for the uploads folder. Typed as an open interface
  `{ id?, upload-type?, title?, location?, [key: string]: unknown }`.

- **`ClientVideoOutput`:** the spec gives no concrete fields. Typed as
  a partial with index signature.

- **`AccessToken.expires`:** the spec hints it may be `null`. Typed as
  `IsoDateTime | null | undefined`.

## 7. Changelog implementation

| § | Item                         | Implementation                                                                             |
| - | ---------------------------- | ------------------------------------------------------------------------------------------ |
| 1 | Cross-canvas clone           | `session.widgets.clone({ sourceCanvasId, sourceWidgetId, destCanvasId, widgetType, ... })`. **NB:** the request body uses underscored keys `source_canvas_id` / `source_widget_id` per the spec — most other API fields use hyphens. |
| 2 | IP Video / RDP POST omitted  | `ipVideos`/`rdpConnections` namespaces have list/get/update/delete only — no `create`.     |
| 3 | Table `column-widths/row-heights` removed | Not in the `Table` type at all.                                               |
| 4 | Table `grid-size` PATCH ignored | `tables.update()` strips `grid-size` from the body and emits `logger.warn`.             |
| 5 | RDP field-naming uncertainty | Flagged in the `RdpConnection` type comment; underscored variant chosen for v0.1.          |

## 8. Coverage

| Group        | Spec'd | Implemented | Notes                                                                   |
| ------------ | -----: | ----------: | ----------------------------------------------------------------------- |
| Canvases     |     17 |          17 | Including 5 subscribe variants                                          |
| Widgets      |     64 |          62 | IP Video POST and RDP POST omitted per changelog §2                     |
| Auth         |     13 |          13 |                                                                         |
| Users        |     19 |          19 |                                                                         |
| Folders      |     12 |          12 | Both POST and PATCH variants of `/move` and `/copy`                     |
| Assets       |      3 |           3 |                                                                         |
| Server       |     22 |          22 |                                                                         |
| **Total**    | **150**| **148**     | The 2 missing are deliberate per changelog §2; SDK is API-complete.     |

## 9. Top 3 risks for morning review

1. **RDP / IP-Video field naming** (changelog §5) — unverified. Half a
   day on a live server clears this.
2. **Workspace / VideoOutput response shapes** — typed defensively as
   open interfaces. We should grab a real response and tighten them
   before declaring v1.
3. **No camelCase conversion** — intentional, but Jaypaul may prefer
   the more idiomatic JS style. Decision-class: easy to change in v0.2
   if we add a normaliser, hard to change later if downstream code
   indexes by `canvas["canvas-name"]` everywhere. Worth a 5-minute
   conversation.
