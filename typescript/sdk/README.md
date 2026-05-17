# @mt-canvus-tools/sdk

TypeScript SDK for the Canvus REST API.

> **Status:** v0.1.0 — greenfield. Covers all 150 documented endpoints
> across canvases, widgets, auth, users, folders, assets, and server admin.
> See `IMPLEMENTATION-NOTES.md` for the autonomous design decisions made
> during the initial scaffold.

## Install

```sh
pnpm add @mt-canvus-tools/sdk
```

Requires Node 20+ (built-in `fetch`, Web Streams).

## Quick start

```ts
import { createSession } from "@mt-canvus-tools/sdk";

const session = createSession({
  baseUrl: "https://canvus.example.com/api/v1/",
  apiKey: process.env.CANVUS_API_KEY,
});

for (const canvas of await session.canvases.list()) {
  console.log(canvas["canvas-name"]);
}
```

The session exposes one resource namespace per documented endpoint group:

| Namespace          | Endpoints |
| ------------------ | --------- |
| `session.canvases` | 17        |
| `session.widgets`  | 64        |
| `session.auth`     | 13        |
| `session.users`    | 19        |
| `session.folders`  | 12        |
| `session.assets`   | 3         |
| `session.server`   | 22        |

## Streaming

Every list/detail endpoint that supports `?subscribe=true` is exposed as
an async iterator backed by `undici` for true backpressure:

```ts
const ctrl = new AbortController();
for await (const canvas of session.canvases.subscribe({ signal: ctrl.signal })) {
  console.log("canvas event:", canvas["canvas-name"]);
}
```

Break out of the loop or abort the signal to tear down the connection.

## Errors

All thrown errors extend `CanvusError` and carry a discriminator field
`kind` for exhaustive handling:

```ts
import { APIError, AuthError, NotFoundError } from "@mt-canvus-tools/sdk";

try {
  await session.canvases.get("missing");
} catch (err) {
  if (err instanceof NotFoundError) return null;
  if (err instanceof AuthError) await refreshToken();
  throw err;
}
```

## Conventions

- Wire-shape preserved exactly: the SDK returns kebab-case JSON keys as
  the server emits them (no camelCase conversion).
- Named exports only; no default exports.
- Pixel coordinates throughout (never normalised 0–1).
- See `docs/conventions/typescript.md` in the repo root for the locked
  TypeScript conventions every package in this workspace follows.

## Scripts

```sh
pnpm build            # tsup → dual ESM/CJS in dist/
pnpm test             # vitest unit tests with v8 coverage
pnpm test:integration # live-server tests (requires CANVUS_API_KEY env)
pnpm lint
pnpm typecheck
```

## License

MIT — see `LICENSE`.
