# autolayout

An auto-layout zone for Canvus. Drop widgets (screen shares, videos, images,
PDFs, notes, browsers) into a named anchor and they arrange themselves around
a main source, the way a Mersive Solstice grid or a T1V quad does, but with
five modes, aspect-aware tiling and drag gestures for choosing the main source.

Canvus has no plugin API for new native widget types, so the "widget" is an
anchor plus this service: it watches the canvas over the REST API and moves
and scales the widgets inside each zone. Nothing on the clients changes.

## Use it

1. In Canvus, create an anchor and name it `Auto Layout` (or `Auto Layout: large left`).
2. Run the service against the canvas:

```bash
cd go/tools/autolayout
CANVUS_API_URL=https://server/api/v1 CANVUS_API_KEY=... go run ./cmd/autolayout run --canvas <canvas-id>
```

3. Drag widgets into the anchor. Rename the anchor to change mode.

| Anchor name | Mode |
|---|---|
| `Auto Layout` / `Auto Layout: auto` | Auto (see below) |
| `Auto Layout: large left` | Main left, others in a strip on the right |
| `Auto Layout: large right` | Main right, strip on the left |
| `Auto Layout: large center` (or `centre`) | Main in the middle, others alternate left and right |
| `Auto Layout: grid` (or `even grid`) | Equal cells, main in the first (top-left) cell |

Flags: `--interval` (poll, default 400ms), `--settle` (stillness before acting,
default 600ms), `--note-padding` (30 for the legacy Note model, 0 normalised),
`--dry-run`, `--once`.

## Gestures

| Do this | Result |
|---|---|
| Drop a widget in the zone | Joins at the end of the order |
| Drop it on the large spot (large modes) | Joins as the main source |
| Drag a member onto the main slot | It becomes main; the old main takes its slot |
| Drag a member onto another tile | The two swap |
| Drag a member out of the zone | It leaves; the rest re-flow, order kept |
| Let go anywhere else in the zone | It snaps back |
| Pin a widget | It is left alone |

The service never moves a widget that is still moving (it waits for `--settle`),
and it recognises its own writes, so it does not fight people.

## Auto

* One source fills the zone (every mode does this, as Solstice does).
* Two or more, nobody promoted: even grid.
* Exactly one live source (video input, IP video, RDP, browser) among static
  content: that source becomes main and gets the large spot.
* Someone promoted a main (dragged it onto the first tile or the large spot):
  large left, or large center on a wide zone (aspect 2:1 or more) with four or
  more secondaries.
* If the large layout would make any secondary smaller than 2% of the zone,
  Auto falls back to the grid.

## Layout rules

* No cropping or stretching: each source is fitted inside its cell with its
  aspect ratio kept, centred.
* Large modes give the main source its natural width at full zone height,
  clamped so it takes at least half the width (40% for centre) and each strip
  at least 18%.
* Grids pick the column count that shows the most content for the actual
  source aspects (four 16:9 sources make a 2x2; three phones make a row). A
  short last row is centred.
* Order is arrival order and is kept on add and remove, so tiles never shuffle.
* Gaps and margins scale with the zone (1.5% of its short side).

## Preview without a server

```bash
go run ./cmd/autolayout preview --mode large-center --zone 5760x1080 --aspects 16:9,16:9,16:9,4:3,9:16
```

prints the plan as JSON.

## Limits of this prototype

* Polls the REST API; the subscribe stream would cut latency.
* Moves are single PATCHes; `layout.Lerp` is there for animated moves but the
  runner does not tween yet.
* Only root-level widgets are arranged; children of groups/other widgets are not.
* Tested against the layout and controller logic and the geometry round trip,
  not yet against a live Canvus server.
