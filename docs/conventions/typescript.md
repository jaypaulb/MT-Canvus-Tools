# TypeScript Conventions — MT-Canvus-Tools

**Status:** Locked defaults for Phase 4b autonomous refresh.
**Scope:** All TypeScript code under `typescript/` (SDK, examples, tools).
**Audience:** Humans and coding agents authoring or refreshing TypeScript artifacts in this monorepo.

This document fixes the build, lint, log, error, test, and CI defaults for every TypeScript package in the repo. Phase 4b refresh agents follow it without prompting Jaypaul. Deviations require an entry in the [Convention amendments](#convention-amendments) section at the bottom of this file.

> **Note on greenfield status.** Unlike the Go and Python SDKs, which were ported from existing repositories, the TypeScript SDK is being written from scratch against `docs/api-reference/`. There is therefore no legacy code to honour — every choice below is the canonical choice for the new SDK, the eight core examples, and the rewritten `examples/webui`.

---

## Module structure

The TypeScript tree is a **single pnpm workspace** rooted at `typescript/`. The workspace owns the lockfile, the shared tooling configs, and the cross-package scripts. Each package is independently versioned and independently publishable, but they are developed and tested together.

```
typescript/
├── package.json                 # workspace root (private: true)
├── pnpm-workspace.yaml          # workspace member globs
├── pnpm-lock.yaml               # single lockfile for all packages
├── tsconfig.base.json           # shared compiler options
├── eslint.config.mjs            # flat config, applied workspace-wide
├── prettier.config.mjs
├── .husky/                      # pre-commit hooks
├── sdk/                         # @mt-canvus-tools/sdk
│   ├── package.json
│   ├── tsconfig.json
│   ├── src/
│   └── tests/
│       └── integration/
├── examples/
│   ├── core/
│   │   ├── 01-auth-and-list/    # @mt-canvus-tools/example-auth-and-list
│   │   ├── 02-auth-flows/
│   │   ├── 03-widget-crud/
│   │   ├── 04-file-upload/
│   │   ├── 05-streaming/
│   │   ├── 06-llm-integration/
│   │   ├── 07-webhooks-notifications/
│   │   └── 08-cross-canvas-clone/
│   └── webui/                   # @mt-canvus-tools/example-webui
```

`pnpm-workspace.yaml`:

```yaml
packages:
  - "sdk"
  - "examples/core/*"
  - "examples/webui"
```

### Package naming

- SDK: `@mt-canvus-tools/sdk`
- Core examples: `@mt-canvus-tools/example-<slug>` (e.g. `@mt-canvus-tools/example-auth-and-list`)
- Web UI example: `@mt-canvus-tools/example-webui`

The `@mt-canvus-tools/` scope is reserved on npm. Only the SDK ships to npm in v1; examples are `private: true` and exist solely to be cloned, read, or run locally.

Examples consume the SDK as a workspace dependency:

```json
{
  "dependencies": {
    "@mt-canvus-tools/sdk": "workspace:*"
  }
}
```

This guarantees examples are always exercising the SDK source in the same repo, not a stale published version.

---

## Node version

- **Minimum runtime:** Node 20 LTS (the first LTS with stable `fetch`, `Web Streams`, `AbortSignal.timeout`, and `--env-file`).
- **CI matrix:** Node 20 and Node 22. Node 22 is not the minimum because some downstream consumers still pin to 20, but we guard against regressions on it.
- **No Node 18.** It is EOL as of April 2025 and lacks the `fetch` stability we rely on.

Set this explicitly in every `package.json`:

```json
{
  "engines": {
    "node": ">=20.0.0"
  }
}
```

The SDK is intended to also work in the browser and in Deno (Deno 1.40+). Browser/Deno are not in the CI matrix in v1 but the SDK code is written to avoid Node-only APIs except in clearly marked submodules (`@mt-canvus-tools/sdk/node`).

---

## Package manager

- **pnpm 9 or newer.** No npm, no yarn.
- A single `pnpm-lock.yaml` at `typescript/`. Per-package lockfiles are an error.
- `packageManager` is pinned at the workspace root so `corepack` picks the right version:

```json
{
  "packageManager": "pnpm@9.12.0"
}
```

Common workspace scripts (run from `typescript/`):

```bash
pnpm install                       # install everything, frozen-ish in CI
pnpm -r build                      # build every package in topological order
pnpm -r test                       # run vitest in every package
pnpm -r lint
pnpm --filter @mt-canvus-tools/sdk test
pnpm --filter "./examples/core/*" build
```

Rationale: pnpm gives us symlinked workspace dependencies (so SDK edits propagate to examples instantly), the strictest dependency resolution of the three managers, and the fastest CI installs via its content-addressable store.

---

## Build

- **Compiler:** `tsc` with project references, used for typechecking and for examples.
- **SDK bundler:** `tsup` (esbuild under the hood), used to emit the dual ESM/CJS artefact published to npm.
- **Output:** every package writes to its own `dist/`. `dist/` is gitignored.

### Project references

The workspace root holds a "solution" `tsconfig.json` that references every member package. Each member's `tsconfig.json` extends `tsconfig.base.json` and declares references to anything it consumes (e.g. examples reference the SDK). This gives:

- Incremental rebuilds (only changed packages and their dependents are touched).
- A single `pnpm -r tsc -b` command at the root that builds everything in the right order.
- Editor-friendly cross-package go-to-definition into source, not into `dist/`.

`typescript/tsconfig.json` (solution file):

```json
{
  "files": [],
  "references": [
    { "path": "./sdk" },
    { "path": "./examples/core/01-auth-and-list" },
    { "path": "./examples/core/02-auth-flows" },
    { "path": "./examples/core/03-widget-crud" },
    { "path": "./examples/core/04-file-upload" },
    { "path": "./examples/core/05-streaming" },
    { "path": "./examples/core/06-llm-integration" },
    { "path": "./examples/core/07-webhooks-notifications" },
    { "path": "./examples/core/08-cross-canvas-clone" },
    { "path": "./examples/webui" }
  ]
}
```

### SDK dual-publish via tsup

The SDK ships ESM as the primary format and CJS as a compatibility shim. Modern Node and bundlers will resolve ESM; legacy CJS consumers still work.

`sdk/tsup.config.ts`:

```ts
import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts", "src/node/index.ts"],
  format: ["esm", "cjs"],
  dts: true,
  sourcemap: true,
  clean: true,
  target: "node20",
  splitting: false,
  treeshake: true,
});
```

The `package.json` declares both formats via the `exports` map. Never rely on the legacy `main` field alone — modern bundlers ignore it.

```json
{
  "name": "@mt-canvus-tools/sdk",
  "type": "module",
  "main": "./dist/index.cjs",
  "module": "./dist/index.js",
  "types": "./dist/index.d.ts",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js",
      "require": "./dist/index.cjs"
    },
    "./node": {
      "types": "./dist/node/index.d.ts",
      "import": "./dist/node/index.js",
      "require": "./dist/node/index.cjs"
    }
  },
  "files": ["dist", "README.md", "LICENSE"]
}
```

Examples do **not** use tsup. They are run directly with `tsx` in development (`tsx src/index.ts`) and built with `tsc` for distribution.

---

## TS config

A single base config at `typescript/tsconfig.base.json` defines the strict defaults. Every package extends it.

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "lib": ["ES2022"],
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "noFallthroughCasesInSwitch": true,
    "exactOptionalPropertyTypes": true,
    "verbatimModuleSyntax": true,
    "isolatedModules": true,
    "esModuleInterop": true,
    "forceConsistentCasingInFileNames": true,
    "skipLibCheck": true,
    "resolveJsonModule": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "composite": true,
    "incremental": true
  }
}
```

Per-package `tsconfig.json` extends this and sets paths:

```json
{
  "extends": "../tsconfig.base.json",
  "compilerOptions": {
    "outDir": "./dist",
    "rootDir": "./src",
    "tsBuildInfoFile": "./dist/.tsbuildinfo"
  },
  "include": ["src/**/*"],
  "exclude": ["dist", "node_modules", "**/*.test.ts"],
  "references": [{ "path": "../../sdk" }]
}
```

### Why each strict flag

- `noUncheckedIndexedAccess`: forces explicit handling of `T | undefined` when indexing arrays/records. Catches a whole class of "off-by-one returns `undefined`" bugs at compile time. Yes, it is noisy. That is the point.
- `exactOptionalPropertyTypes`: distinguishes `{ x?: number }` (omitted) from `{ x: number | undefined }` (set to undefined). Matters when serialising to JSON for the Canvus API.
- `verbatimModuleSyntax`: requires `import type` for type-only imports, which keeps the emitted output clean and makes the boundary between values and types unambiguous.
- `isolatedModules`: ensures every file can be compiled in isolation, which tsup/esbuild require.

---

## Linting

- **ESLint v9** with the flat config (`eslint.config.mjs`).
- **`@typescript-eslint`** with the strict-type-checked preset.
- **Prettier** for formatting, run separately and via lint-staged.
- **Husky + lint-staged** for pre-commit enforcement.

### Flat ESLint config

`typescript/eslint.config.mjs`:

```js
import eslint from "@eslint/js";
import tseslint from "typescript-eslint";
import prettier from "eslint-config-prettier";

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.strictTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,
  {
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      "@typescript-eslint/consistent-type-imports": "error",
      "@typescript-eslint/no-floating-promises": "error",
      "@typescript-eslint/no-misused-promises": "error",
      "@typescript-eslint/require-await": "error",
      "@typescript-eslint/switch-exhaustiveness-check": "error",
      "no-console": ["error", { allow: ["warn", "error"] }],
    },
  },
  {
    files: ["**/*.test.ts", "**/tests/**/*.ts"],
    rules: {
      "@typescript-eslint/no-non-null-assertion": "off",
    },
  },
  prettier,
);
```

### Prettier

Single config at `typescript/prettier.config.mjs`:

```js
export default {
  semi: true,
  singleQuote: false,
  trailingComma: "all",
  printWidth: 100,
  arrowParens: "always",
};
```

Prettier is the last word on whitespace. ESLint stylistic rules that conflict with Prettier are disabled via `eslint-config-prettier` (last in the config array).

### Pre-commit

`typescript/.husky/pre-commit`:

```sh
pnpm lint-staged
```

`typescript/package.json`:

```json
{
  "lint-staged": {
    "*.{ts,tsx,mjs,cjs,js}": ["eslint --fix", "prettier --write"],
    "*.{json,md,yml,yaml}": ["prettier --write"]
  }
}
```

Pre-commit failures block the commit. Agents that hit a pre-commit failure must fix the underlying issue and create a new commit (see global protocol — never `--amend` past a failed hook).

---

## Logging

- **Library:** `pino`.
- **Dev transport:** `pino-pretty` (human-readable, coloured).
- **Prod transport:** raw JSON to stdout.
- **Selection driver:** the `LOG_FORMAT` env var (`json` or `pretty`).

### Single configured instance per package

Each package owns one configured logger module. Code imports the logger from there, never constructs new pino instances inline.

`sdk/src/logging.ts`:

```ts
import pino from "pino";

const isPretty = process.env.LOG_FORMAT === "pretty";

export const logger = pino({
  level: process.env.LOG_LEVEL ?? "info",
  base: { component: "sdk" },
  transport: isPretty
    ? {
        target: "pino-pretty",
        options: { colorize: true, translateTime: "SYS:HH:MM:ss.l" },
      }
    : undefined,
});

export type Logger = typeof logger;
```

Consumers:

```ts
import { logger } from "./logging";

logger.info({ canvasId }, "fetched canvas");
logger.error({ err, canvasId }, "canvas fetch failed");
```

### Browser

In the browser build, `pino` falls back to `console`-style output and the transport block is dropped (it depends on Node `worker_threads`). The `examples/webui` package may wrap this with a thin facade that also surfaces errors in a toast component, but the underlying logger is still pino.

### Why pino

- Structured JSON is the only sane choice for shipping logs to anywhere beyond a terminal.
- It is the fastest mainstream logger; the SDK runs inside MCP servers and webhook handlers where allocation pressure matters.
- The `pino-pretty` dev transport gives the human-readable output people actually want locally, without forcing a different logger in prod.

### Forbidden

- `console.log` in library or production code. ESLint enforces this. `console.warn` and `console.error` are allowed for genuine emergencies (e.g. a logger that fails to initialise) and for example projects where readability beats discipline.
- `winston`, `bunyan`, `loglevel`, custom log facades. One logger, one config path.

---

## Error handling

The SDK defines a small, exhaustive hierarchy of error classes. Every error thrown from SDK code is an instance of `CanvusError`. Consumers can `instanceof`-narrow to the subclass they care about, or switch on the discriminator field for total-function handling.

### Hierarchy

```ts
// sdk/src/errors.ts

export type CanvusErrorKind =
  | "api"
  | "validation"
  | "auth"
  | "network";

export class CanvusError extends Error {
  public readonly kind: CanvusErrorKind;

  constructor(kind: CanvusErrorKind, message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "CanvusError";
    this.kind = kind;
  }
}

export class APIError extends CanvusError {
  constructor(
    public readonly status: number,
    public readonly body: unknown,
    message: string,
    options?: ErrorOptions,
  ) {
    super("api", message, options);
    this.name = "APIError";
  }
}

export class ValidationError extends CanvusError {
  constructor(
    public readonly issues: readonly ValidationIssue[],
    message: string,
    options?: ErrorOptions,
  ) {
    super("validation", message, options);
    this.name = "ValidationError";
  }
}

export class AuthError extends CanvusError {
  constructor(
    public readonly reason: "missing-token" | "expired" | "forbidden",
    message: string,
    options?: ErrorOptions,
  ) {
    super("auth", message, options);
    this.name = "AuthError";
  }
}

export class NetworkError extends CanvusError {
  constructor(message: string, options?: ErrorOptions) {
    super("network", message, options);
    this.name = "NetworkError";
  }
}

export interface ValidationIssue {
  readonly path: ReadonlyArray<string | number>;
  readonly message: string;
}
```

### `cause` chains

Always preserve the underlying error via the standard `Error.cause` mechanism, never by string concatenation.

```ts
try {
  await fetch(url);
} catch (err) {
  throw new NetworkError(`request to ${url} failed`, { cause: err });
}
```

### Discriminated-union handling

For callers that need to branch on every possible failure mode, the `kind` discriminator gives exhaustive checking when combined with `noFallthroughCasesInSwitch`:

```ts
function describe(err: CanvusError): string {
  switch (err.kind) {
    case "api":
      return `server returned ${(err as APIError).status}`;
    case "validation":
      return `bad input: ${(err as ValidationError).issues.length} issue(s)`;
    case "auth":
      return `auth failed: ${(err as AuthError).reason}`;
    case "network":
      return "network unreachable";
  }
}
```

### No silent fallbacks

Per `~/.claude/CLAUDE.md`'s "On Fallbacks" rule: never swallow an error and return a default. If the operation failed, propagate. If the caller wants a fallback, they wrap in `try` themselves.

```ts
// WRONG — silent corruption
const canvas = await session.getCanvas(id).catch(() => null);

// RIGHT — fail loudly, caller decides
const canvas = await session.getCanvas(id);
```

---

## Configuration

All runtime configuration is read from environment variables, parsed once at startup via a `zod` schema, and frozen into a typed object.

### Env-var prefix

All Canvus-related env vars use the `CANVUS_` prefix. Examples:

| Var                    | Purpose                                         |
| ---------------------- | ----------------------------------------------- |
| `CANVUS_API_BASE_URL`  | e.g. `https://canvus.example.com/api/v1/`       |
| `CANVUS_API_KEY`       | Long-lived API key (`Private-Token` header)     |
| `CANVUS_TIMEOUT_MS`    | Request timeout in milliseconds                 |
| `CANVUS_VERIFY_TLS`    | `true`/`false` — useful only for dev with self-signed certs |
| `LOG_LEVEL`            | pino level (not prefixed; standard convention)  |
| `LOG_FORMAT`           | `json` or `pretty` (not prefixed)               |

### `loadConfig()`

```ts
// sdk/src/config.ts
import { z } from "zod";

const ConfigSchema = z.object({
  apiBaseUrl: z.string().url(),
  apiKey: z.string().min(1).optional(),
  timeoutMs: z.coerce.number().int().positive().default(30_000),
  verifyTls: z
    .enum(["true", "false"])
    .default("true")
    .transform((v) => v === "true"),
});

export type Config = z.infer<typeof ConfigSchema>;

export function loadConfig(env: NodeJS.ProcessEnv = process.env): Config {
  const parsed = ConfigSchema.safeParse({
    apiBaseUrl: env.CANVUS_API_BASE_URL,
    apiKey: env.CANVUS_API_KEY,
    timeoutMs: env.CANVUS_TIMEOUT_MS,
    verifyTls: env.CANVUS_VERIFY_TLS,
  });

  if (!parsed.success) {
    throw new ValidationError(
      parsed.error.issues.map((i) => ({ path: i.path, message: i.message })),
      "invalid CANVUS_* configuration",
    );
  }
  return Object.freeze(parsed.data);
}
```

Always parse once at startup. Never read `process.env` deep inside business logic — that makes the code untestable and surprises consumers when they thought they had overridden something.

For tests, pass an explicit env object: `loadConfig({ CANVUS_API_BASE_URL: "...", ... })`. Vitest does not need `dotenv` because Node 20 supports `--env-file=.env` natively.

---

## HTTP client

### Base case: native `fetch`

The SDK uses Node 20's built-in `fetch` for ordinary request/response calls. No `axios`, no `node-fetch`, no `got`. One stack, one error surface, no extra dependency.

```ts
// sdk/src/http.ts
export async function request<T>(
  config: Config,
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  const url = new URL(path, config.apiBaseUrl);
  const headers = new Headers({ "content-type": "application/json" });
  if (config.apiKey) headers.set("private-token", config.apiKey);

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), config.timeoutMs);

  let response: Response;
  try {
    response = await fetch(url, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
      signal: controller.signal,
    });
  } catch (err) {
    throw new NetworkError(`${method} ${url.pathname} failed`, { cause: err });
  } finally {
    clearTimeout(timeout);
  }

  if (!response.ok) {
    const text = await response.text();
    throw new APIError(response.status, safeJson(text), `${method} ${url.pathname} returned ${response.status}`);
  }

  return (await response.json()) as T;
}
```

### Streaming: `undici` Dispatcher

The Canvus API supports HTTPS streaming via `?subscribe=true`, returning newline-delimited JSON. `fetch` can read these via `response.body`, but for true backpressure, connection reuse, and finer control over chunk timing we use `undici` directly via a `Dispatcher`.

Streams are exposed to consumers as **async iterators**:

```ts
// usage
for await (const event of session.subscribeCanvas(canvasId)) {
  if (event.kind === "widget-updated") {
    logger.info({ widgetId: event.widget.id }, "widget changed");
  }
}
```

```ts
// sdk/src/streaming.ts
import { Client, type Dispatcher } from "undici";

export async function* streamNDJSON<T>(
  client: Dispatcher,
  path: string,
  signal: AbortSignal,
): AsyncGenerator<T, void, void> {
  const { body, statusCode } = await client.request({
    path,
    method: "GET",
    signal,
  });

  if (statusCode >= 400) {
    throw new APIError(statusCode, await body.text(), `stream ${path} failed`);
  }

  let buffer = "";
  for await (const chunk of body) {
    buffer += chunk.toString("utf8");
    let newlineIdx: number;
    while ((newlineIdx = buffer.indexOf("\n")) !== -1) {
      const line = buffer.slice(0, newlineIdx).trim();
      buffer = buffer.slice(newlineIdx + 1);
      if (line) yield JSON.parse(line) as T;
    }
  }
}
```

Consumers cancel a stream by aborting the `AbortSignal` they passed in. Breaking out of the `for await` loop with `break`/`return` will also dispose the iterator and close the underlying connection thanks to the generator semantics.

### Why two stacks

`fetch` is ergonomic for request/response. `undici`'s `Dispatcher` is the right tool for long-lived streams: keep-alive, configurable headers timeout, no implicit buffering of the full body. Keeping both is cheaper than half-using one.

---

## Testing

- **Runner:** `vitest`.
- **Coverage:** `@vitest/coverage-v8`.
- **Target coverage:** **70%** for the SDK (lines + branches). Examples are not coverage-gated.
- **Test file location:** `*.test.ts` co-located with source (`src/canvas.ts` ↔ `src/canvas.test.ts`).
- **Integration tests:** isolated under `sdk/tests/integration/` with their own config.

### Why vitest

It is the only mainstream runner that takes our TS config at face value: native ESM, no Jest-style transform pipeline, native `import.meta`, parallel by default, the same expect API people already know. It also gives us in-source tests for tiny utilities if we ever want them.

### Unit-test config

`sdk/vitest.config.ts`:

```ts
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: ["src/**/*.test.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
      thresholds: { lines: 70, branches: 70, functions: 70, statements: 70 },
      include: ["src/**/*.ts"],
      exclude: ["src/**/*.test.ts", "src/**/index.ts"],
    },
  },
});
```

### Integration-test config

Integration tests talk to a real Canvus server (or a mock one stood up via `msw`). They are gated behind a separate command and a `settings.json` file (the same pattern used by the Go SDK):

`sdk/vitest.integration.config.ts`:

```ts
import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: ["tests/integration/**/*.test.ts"],
    testTimeout: 30_000,
    hookTimeout: 30_000,
    coverage: { enabled: false },
  },
});
```

`sdk/package.json` scripts:

```json
{
  "scripts": {
    "test": "vitest run --coverage",
    "test:watch": "vitest",
    "test:integration": "vitest run --config vitest.integration.config.ts"
  }
}
```

CI runs `pnpm test` always. `pnpm test:integration` runs only on a labelled job with a Canvus server URL injected via secrets.

### Test style

- One `describe` block per exported symbol.
- Arrange / Act / Assert visually separated by blank lines.
- Prefer `expect.toEqual` over deep snapshots; snapshots are for stable serialised output only.
- Mock at the HTTP boundary (with `msw` or `undici`'s `MockAgent`), never at the SDK method boundary. Mocking your own code teaches you nothing.

---

## CI

A single workflow at `.github/workflows/typescript.yml` covers the TypeScript packages. It runs on PRs that touch `typescript/**` and on pushes to `main`.

```yaml
name: typescript

on:
  push:
    branches: [main]
    paths: ["typescript/**", ".github/workflows/typescript.yml"]
  pull_request:
    paths: ["typescript/**", ".github/workflows/typescript.yml"]

defaults:
  run:
    working-directory: typescript

jobs:
  ci:
    strategy:
      fail-fast: false
      matrix:
        node-version: [20, 22]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: pnpm/action-setup@v4
        with: { version: 9 }

      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
          cache: pnpm
          cache-dependency-path: typescript/pnpm-lock.yaml

      - run: pnpm install --frozen-lockfile

      - run: pnpm -r lint
      - run: pnpm -r typecheck
      - run: pnpm -r test
      - run: pnpm -r build
```

Notes:

- `--frozen-lockfile` ensures CI fails if the lockfile is out of sync; never regenerate it silently.
- `pnpm -r` runs each script in every workspace package in topological order.
- Coverage reports are uploaded via a separate step on Node 20 only (avoid double-uploads).
- Integration tests live in a manually-triggered workflow (`typescript-integration.yml`) and require a `CANVUS_API_BASE_URL` + `CANVUS_API_KEY` secret pair.

---

## Code style

### Exports

- **Named exports only.** No `export default`.
- Re-export the public API from `src/index.ts`. Anything not re-exported there is internal and may change without a major version bump.

```ts
// GOOD
export class Session { /* ... */ }

// BAD
export default class Session { /* ... */ }
```

Rationale: default exports rename freely at the import site, defeating grep, IDE rename-symbol, and consistent docs. Named exports give every symbol exactly one canonical name.

### File naming

- **kebab-case** for files and directories: `canvas-session.ts`, `widget-types.ts`, `auth/api-key-strategy.ts`.
- Test files mirror their subject: `canvas-session.test.ts`.
- PascalCase for the symbols inside: `class CanvasSession`, `interface WidgetType`.

### Interfaces vs types

- `interface` for **object shapes that may be extended** (e.g. widget schemas, public API surfaces).
- `type` for **unions, intersections, mapped/conditional types, and anything that is not an extendable object shape**.

```ts
// Object shape consumers might extend — interface
export interface Widget {
  id: string;
  kind: WidgetKind;
}

// Union — type
export type WidgetKind = "note" | "image" | "video" | "browser";

// Mapped — type
export type PartialWidget = Partial<Widget>;
```

### Public API docs

Every exported symbol — class, function, type, interface, const — must carry a TSDoc comment. The first line is a one-sentence summary. Use `@param`, `@returns`, `@throws`, `@example` where they earn their keep.

```ts
/**
 * Lists every canvas accessible to the authenticated user.
 *
 * @returns Array of canvas metadata, ordered by server-side default (most recent first).
 * @throws {AuthError} If the configured API key is missing or rejected.
 * @throws {NetworkError} On transport-level failure.
 * @example
 * ```ts
 * const session = new Session(loadConfig());
 * for (const c of await session.listCanvases()) {
 *   console.log(c.name);
 * }
 * ```
 */
export async function listCanvases(): Promise<Canvas[]> { /* ... */ }
```

This is non-negotiable for the SDK. TypeDoc is run as part of the docs build and treats missing comments as warnings.

### Async / Promises

- `async`/`await` everywhere. No raw `.then()` chains in library code.
- Never leave a floating promise (`@typescript-eslint/no-floating-promises` enforces this).
- Use `Promise.all` for parallel independent work; `Promise.allSettled` when partial failure should not abort.

### Imports

- Type-only imports use `import type` (enforced by `verbatimModuleSyntax`).
- Order: node built-ins → third-party → workspace packages → relative. Prettier + an import-sort plugin handles this mechanically.
- No barrel-only deep imports across package boundaries — always import from the package's documented entry point.

### Nullability

- Prefer `undefined` over `null` for "no value." JSON serialisation occasionally requires `null`; explicitly use it there.
- Don't `!` away `noUncheckedIndexedAccess`. Guard properly.

```ts
// BAD
const first = list[0]!;

// GOOD
const first = list[0];
if (first === undefined) throw new ValidationError([], "list is empty");
```

---

## Convention amendments

Phase 4b refresh agents append to this section whenever they deviate from the locked defaults. Format:

### YYYY-MM-DD — <one-line summary>

**Item:** <tool or example being refreshed>
**Decision:** <what was chosen>
**Rationale:** <why>

<!-- No amendments yet. -->
