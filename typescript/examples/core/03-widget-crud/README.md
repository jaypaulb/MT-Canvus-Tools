# Example 03 — widget-crud

Walks the create / update / delete / verify lifecycle on a single
sticky-note widget.

## Purpose

Demonstrates the canonical CRUD pattern on a typed widget resource
(`session.widgets.notes`), including how to verify a delete by
catching the SDK's `NotFoundError`.

The six steps are:

1. Create a note with the current ISO timestamp in the text.
2. Print the new widget id.
3. PATCH the note: new text + green background (`#3aaa34ff`).
4. Print the updated widget.
5. DELETE the note.
6. Verify the delete by GETting the same id and expecting
   `NotFoundError`.

If any step fails the example stops immediately. There is no point
running steps 2–6 once step 1 has failed.

## Prerequisites

- Node 20+ and pnpm 9+
- An API key with edit access to `CANVUS_CANVAS_ID`

## Configuration

| Variable           | Required | Description                                       |
| ------------------ | -------- | ------------------------------------------------- |
| `CANVUS_API_URL`  | yes      | API base URL                                      |
| `CANVUS_API_KEY`   | yes      | Long-lived API key with edit access               |
| `CANVUS_CANVAS_ID` | yes      | Canvas the note will be written to                |

## Run

```bash
pnpm dev
```

## Expected output

```
{"level":30,"component":"example-widget-crud","noteId":"...","text":"hello from TypeScript @ 2026-05-18T12:34:56.000Z","msg":"created note"}
{"level":30,"component":"example-widget-crud","noteId":"...","text":"updated from TypeScript @ ...","backgroundColor":"#3aaa34ff","msg":"updated note"}
{"level":30,"component":"example-widget-crud","noteId":"...","msg":"deleted note"}
{"level":30,"component":"example-widget-crud","noteId":"...","msg":"delete verified — GET returned 404 as expected"}
```

Note: the SDK Note type uses underscored field names (`id`,
`background_color`, `text_color`, `auto_text_color`) to match the live
Canvus v1.2 wire shape.

## How it works

The SDK exposes typed widget sub-resources on
`session.widgets.<kind>`, where each sub-resource has the same
five-method shape (`list`, `get`, `create`, `update`, `delete`) plus a
streaming `subscribe`. This example uses `notes` but the pattern is
identical for `images`, `videos`, `browsers`, etc.

**Verifying deletion** is one of the few times where a 404 is good
news. The SDK throws `NotFoundError` (an `APIError` subclass) for any
404 response, which is `instanceof`-narrowable separately from generic
`APIError`. That distinction matters because 404 is often normal
control flow (the resource never existed, or was just deleted), and
collapsing it into a generic `APIError` would force the caller to
inspect the status code. See `docs/conventions/typescript.md` for the
full error hierarchy.

```ts
try {
  await session.widgets.notes.get(canvasId, noteId);
  // unexpected success — delete did not take effect
} catch (err) {
  if (err instanceof NotFoundError) {
    // success: the resource is gone
  } else {
    throw err; // some other failure
  }
}
```

## Troubleshooting

- **403 on create** — your API key user has view-only access. Need
  edit.
- **404 on create** — `CANVUS_CANVAS_ID` is wrong.
- **Delete-verify fails (note still exists)** — likely a server-side
  caching delay, or the delete returned a 2xx but did not propagate.
  Re-run; if it still fails, file an issue.
