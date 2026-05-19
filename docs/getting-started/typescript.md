# Getting Started — TypeScript

Get up and running with the MT-Canvus-Tools TypeScript SDK.

## Prerequisites

- **Node.js 20 or later** — verify with `node --version`
- **pnpm** — install with `npm install -g pnpm` or see the [pnpm docs](https://pnpm.io/)
- A running Canvus server and an API key

## Option A: Use the SDK as a dependency

> **Note:** Until the package is published on npm, use Option B or link the SDK via a local path.

```bash
pnpm add @mt-canvus-tools/sdk
```

Create `index.ts`:

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

## Option B: Run examples from the monorepo

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/typescript
pnpm install
pnpm build
```

Set credentials, then run any numbered example:

```bash
export CANVUS_API_URL=https://your-server/api/v1
export CANVUS_API_KEY=your-api-key

node examples/core/01-auth-and-list/dist/index.js
node examples/core/05-streaming/dist/index.js
```

## Environment variables

| Variable | Description |
|---|---|
| `CANVUS_API_URL` | Full base URL including `/api/v1` suffix |
| `CANVUS_API_KEY` | Long-lived API key — sent as `Private-Token` header |

Environment-backed config is available via `loadConfig()` (reads `CANVUS_API_URL` and `CANVUS_API_KEY`):

```typescript
import { Session, loadConfig } from "@mt-canvus-tools/sdk";

const session = new Session(loadConfig());
```

## Authentication modes

```typescript
// API key (recommended for services and automation)
const session = createSession({
  baseUrl: process.env.CANVUS_API_URL!,
  apiKey: process.env.CANVUS_API_KEY,
});

// Username + password (interactive flows; short-lived token)
const session = createSession({ baseUrl: process.env.CANVUS_API_URL! });
await session.auth.login({ email: "user@example.com", password: "password" });
```

## Real-time streaming

```typescript
for await (const canvas of session.canvases.subscribe()) {
  console.log(canvas.id, canvas.name);
}
```

## Next steps

- [TypeScript workspace README](../../typescript/README.md)
- [TypeScript conventions](../conventions/typescript.md)
- [API reference](../api-reference/README.md)
- [All core examples](../../typescript/examples/core/)
