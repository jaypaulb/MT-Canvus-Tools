# Source Freeze

This API reference was extracted from the following source-of-truth commits. Downstream SDK phases reference *this captured state*, not a moving upstream target.

| Source | Path | Branch | SHA |
|---|---|---|---|
| `mt-restapi-client` (C++ canonical impl) | `gl/conan/canvus/mt-restapi-client` | `dev` | `8e1b2103ba55dd63b3b8b4737c69d96d4edd94e0` |
| `mt-restapi-tests` (Go integration tests) | `gl/conan/canvus/mt-restapi-tests` | `feature/r3-broadcast-control-api` | `3d971fe98caba6611ab7ebb09f81f2f2c5a19ae6` |
| `canvus-server` (cross-reference for ambiguity + additional test fixtures) | `gl/conan/canvus/canvus-server` | `dev` | `4ed2d94fd822c40b87cf174f9861cd09175e9035` |

Extracted: 2026-05-17

> **Note:** `mt-restapi-tests` was on a feature branch at extraction time (`feature/r3-broadcast-control-api`) rather than the default branch. Examples derived from it may include in-flight test coverage for the R3 broadcast-control API. Verify against `mt-restapi-client` (canonical impl) when in doubt.

## Re-extraction policy

A future post-v1 phase may re-extract from newer SHAs and produce a diff against this captured spec to surface upstream drift. Until then, all SDKs and tools in this monorepo are pinned to the behaviour documented from these SHAs.
