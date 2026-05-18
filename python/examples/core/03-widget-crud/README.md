# 03 — widget-crud

Exercises the full lifecycle of a note widget:

1. **CREATE** — `POST /canvases/{id}/notes` with text + position + size.
2. Echo the new widget id.
3. **PATCH** — change `text` and `background_color`.
4. Echo the updated widget.
5. **DELETE** — `DELETE /canvases/{id}/notes/{nid}`.
6. **VERIFY DELETE** — `GET` the deleted widget; expect `NotFoundError`.

On any step's failure the example stops immediately so the failing step can
be debugged in isolation.

## Prerequisites

| Env var            | Required | Purpose                                          |
| ------------------ | -------- | ------------------------------------------------ |
| `CANVUS_API_URL`   | yes      | Base URL.                                        |
| `CANVUS_API_KEY`   | yes      | API token with edit access on the target canvas. |
| `CANVUS_CANVAS_ID` | yes      | Target canvas id.                                |
| `LOG_FORMAT`       | no       | `console` or `json`.                             |

## Run

```bash
cd python
python examples/core/03-widget-crud/main.py
```

## Expected output

```
created note id: 9c1d…
updated note: text='updated from python @ 2026-05-18T…' bg=#3aaa34ff
verified deletion of note 9c1d… (NotFoundError as expected)
```

## Notes

- Step 6 is the contract test. The SDK's transport classifies 404 into the
  `NotFoundError` subclass, so catching that distinguishes "really gone"
  from "auth error" or "server hiccup".
- Coordinates and sizes are in pixels (Canvus does not use normalised
  coordinates).
- Colour strings are 8-character hex with alpha — `#RRGGBBAA`.

## Troubleshooting

- **`AuthError` (401/403):** API key is wrong or has no edit access on the
  canvas.
- **`NotFoundError` (404) on step 1:** wrong `CANVUS_CANVAS_ID`.
- **Step 6 succeeds (no error):** the note was not actually deleted —
  probably because the underlying DELETE returned a non-2xx that the SDK
  didn't classify as fatal. Re-run with `LOG_LEVEL=DEBUG` and inspect.
