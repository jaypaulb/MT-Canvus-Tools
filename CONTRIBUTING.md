# Contributing to MT-Canvus-Tools

## Overview

This is a monorepo. Most contributions touch one language subtree (`go/`, `python/`, or `typescript/`). Cross-cutting changes (API reference, conventions) are the exception.

## Before you start

Read the convention doc for the language you're working in. Conventions lock build/lint/log/error/test defaults and must be followed without exception. New convention decisions require an amendment entry appended with rationale.

| Language | Convention doc |
|---|---|
| Go | [`docs/conventions/go.md`](docs/conventions/go.md) |
| Python | [`docs/conventions/python.md`](docs/conventions/python.md) |
| TypeScript | [`docs/conventions/typescript.md`](docs/conventions/typescript.md) |

## Development setup

See the per-language getting-started guides:

- [Go](docs/getting-started/go.md)
- [Python](docs/getting-started/python.md)
- [TypeScript](docs/getting-started/typescript.md)

For per-language contributing workflows (toolchain checks, commit scoping, test patterns): [docs/contributing/](docs/contributing/).

## PR workflow

1. Create a feature branch: `git checkout -b feature/<description>` or `fix/<description>`
2. Write tests before implementation (TDD — red/green/refactor)
3. Ensure toolchain passes cleanly before opening a PR:
   - **Go:** `go build ./...` + `go vet ./...` + `go test ./...` + `gofmt -s -l .` (must print nothing)
   - **Python:** `uv run ruff check .` + `uv run pytest` + `uv run mypy --strict sdk/src` (0 errors)
   - **TypeScript:** `pnpm typecheck` + `pnpm build` + `pnpm test` + `pnpm lint`
4. Commit using conventional commit messages: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`
5. Open a PR against `main`

## Secrets

Never commit secrets. Use a `.secrets` file at the repo root (gitignored) for local credentials. Pass credentials via environment variables at runtime.

## API reference corrections

If you discover a discrepancy between `docs/api-reference/` and the live server, add an entry to `docs/api-reference/VERIFIED-CORRECTIONS.md` with curl evidence. See the upstream doc-fix tracker at [canvus-server#96](https://gitlab.multitaction.com/swrd/conan/canvus/canvus-server/-/work_items/96).
