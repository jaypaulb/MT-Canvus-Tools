# Contributing — TypeScript

## Workspace setup

```bash
git clone https://github.com/jaypaulb/MT-Canvus-Tools.git
cd MT-Canvus-Tools/typescript
pnpm install
pnpm build
```

The `package.json` at `typescript/` is the pnpm workspace root. Members: `sdk`, `examples/core/*`, `examples/webui`.

## Adding a new package

1. Create the package directory under `typescript/`
2. Add it to `typescript/pnpm-workspace.yaml`
3. Run `pnpm install`
4. Tests go in the package's `src/__tests__/` using vitest

## Toolchain checks (all must pass)

```bash
pnpm typecheck     # tsc --noEmit — no type errors
pnpm build         # dist/ artifacts produced cleanly
pnpm test          # All vitest tests pass
pnpm lint          # ESLint — must not exceed current lint floor
```

## Lint floor

The ESLint lint floor is **0 errors** after Phase 5. Do not add new lint errors.

## Commit scope

```
feat(typescript/sdk): add createAnyWithAsset
fix(typescript/examples/webui): handle missing canvas_id in relay handler
```

## Convention reference

[`docs/conventions/typescript.md`](../conventions/typescript.md).
