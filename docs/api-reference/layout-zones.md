# Native layout zones

Implementation for https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/702.
Requires a server advertising the layout-zone runtime, not a particular release number.

`/api/v1/canvases/{canvas_id}/layout-zones` supports GET and POST;
`/{zone_id}` supports GET, PATCH and DELETE. GET supports the existing subscribe option.
The public widget type is `LayoutZone`. Generic widget reads retain its layout fields.

Defaults are `layout_mode: auto`, `slot_order: []`, `main_promoted: false`,
`gap_fraction: 0.015`, `layout_version: 1`, and size 960×540 pixels.
Modes: `auto`, `large_left`, `large_right`, `large_center`, `grid`.
All normal geometry, parent, title and pin fields remain available.
`main_promoted` and `gap_fraction` are configurable. `layout_version` is owned
by the algorithm: create accepts version 1; PATCH may echo the stored version,
not change it. Newer-version stored geometry is not overwritten by this runtime.

`slot_order` is always an array, never an encoded JSON string. A PATCH replaces
it explicitly with every current child ID exactly once; it does not adopt unrelated
widgets. Omit fields you do not want changed. Empty arrays and false values are
meaningful. Layout mode/order, promotion and gap changes are server-confirmed.
DELETE releases children with their displayed world poses and native sizes intact.
A timeout is not success: read current state before retrying.

## Go

```go
zone, err := session.CreateLayoutZone(ctx, canvasID, canvus.LayoutZoneRequest{})
// Handle err before using zone.
updated, err := session.UpdateLayoutZone(ctx, canvasID, zone.ID,
    map[string]any{"layout_mode": "grid", "slot_order": []string{}})
```

Use `ListLayoutZones`, `GetLayoutZone`, `DeleteLayoutZone`,
`SubscribeLayoutZones` and `SubscribeLayoutZone` for the remaining operations.
Generic `CreateWidget`, `UpdateWidget` and `DeleteWidget` accept `LayoutZone`.
CLI updates accept `--layout-mode grid` and `--slot-order '["child-id"]'`.

## Python

```python
zone = await client.widgets.layout_zones.create(canvas_id, {})
zone = await client.widgets.layout_zones.update(
    canvas_id, zone.id, {"layout_mode": "large_left", "slot_order": []}
)
```

The resource also supplies `list`, `get`, `delete`, `subscribe`, `subscribe_one`.
Generic `create_any`, `update_any` and `delete_any` recognise `LayoutZone`.

## TypeScript

```typescript
const zone = await session.widgets.layoutZones.create(canvasId, {});
await session.widgets.layoutZones.update(canvasId, zone.id, {
  layout_mode: "grid", slot_order: [],
});
```

The resource also supplies `list`, `get`, `delete`, `subscribe`, `subscribeOne`.
`Widget` includes the LayoutZone discriminated member; generic dispatch accepts it.

These are SDK transport contracts, not claims of released-server availability.
