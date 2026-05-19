# MT-Canvus-Tools — TypeScript

pnpm workspace containing the SDK and examples for the Canvus platform.

## Structure

| Path | Description |
|---|---|
| [`sdk/`](sdk/) | TypeScript SDK — 147 endpoints, 27 typed async iterators, 9 extras modules |
| [`examples/core/`](examples/core/) | 8 numbered examples (01 auth, 02 auth-flows, 03 widget-CRUD, 04 file-upload, 05 streaming, 06 LLM, 07 webhooks, 08 cross-canvas-clone) |
| [`examples/webui/`](examples/webui/) | Hono-based WebUI with RCU canvas relay and server-sent events |

## Workspace setup

Run from this directory (`typescript/`):

```bash
pnpm install       # Install all workspace dependencies
pnpm build         # Build SDK + all examples (ESM + CJS + DTS)
pnpm test          # Run vitest suite
pnpm typecheck     # Type-check without emitting
pnpm lint          # ESLint
```

## Authentication

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key
```

## Quick example

```typescript
import { createSession } from "@mt-canvus-tools/sdk";

const session = createSession({
  baseUrl: process.env.CANVUS_API_URL!,
  apiKey: process.env.CANVUS_API_KEY,
});

const canvases = await session.canvases.list();
for (const canvas of canvases) {
  console.log(canvas.id, canvas.name);
}
```

Or using environment-backed config:

```typescript
import { Session, loadConfig } from "@mt-canvus-tools/sdk";

const session = new Session(loadConfig());
```

## Conventions

[`docs/conventions/typescript.md`](../docs/conventions/typescript.md) (monorepo root).

## Getting started

[`docs/getting-started/typescript.md`](../docs/getting-started/typescript.md) — step-by-step setup, authentication, running examples.
