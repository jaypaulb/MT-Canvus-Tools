# Contributing — Go

## Workspace setup

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/go
go build ./...
go test ./...
```

The `go.work` file at `go/` spans all Go modules. Workspace-level `go build ./...` and `go test ./...` operate across all of them.

## Adding a new package

- Place it under the appropriate module (`sdk/canvus/`, `tools/<name>/internal/`, etc.)
- File size limits: atoms < 50 LOC, molecules < 150 LOC, organisms < 400 LOC (see [`docs/conventions/go.md`](../conventions/go.md) §2)
- Unit tests go in `_test.go` files in the same package
- Integration tests go under `integration/` with the `//go:build integration` tag

## Toolchain checks (all must pass)

```bash
go build ./...           # No build errors
go vet ./...             # No vet warnings
go test ./...            # All tests pass
gofmt -s -l .            # Must print nothing
```

## Commit scope

```
feat(go/sdk): add SubscribeWidgets helper
fix(go/tools/powertoys): guard m.server with mutex
refactor(go/cli): extract auth command into sub-package
```

## Convention reference

[`docs/conventions/go.md`](../conventions/go.md) — locked defaults for error wrapping, logging, context propagation, naming, and file size.
