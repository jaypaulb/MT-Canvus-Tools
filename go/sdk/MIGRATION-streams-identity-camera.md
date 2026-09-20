# Stream, identity and camera migration (SDK-D–H)

This Go-only follow-through preserves the previous actor/retry/outcome safety
contracts. It adds explicit state and geometry boundaries; it does not implement
product permission policy, physical-operator detection, or a new native API.

## Subscriptions

Use `Subscribe[T](ctx, session, relativeEndpoint)` for observable lifecycle and
field presence. For example, the endpoint for workspaces is
`clients/<client-id>/workspaces`; do not supply an absolute URL or query string.

- `Subscription.Frames` retains each object/array boundary, including empty arrays.
  Each `StreamItem.Value` is decoded data; `Raw` retains absent, null, zero,
  unknown and deletion-marker fields. Never treat a zero-valued DTO as a delta.
- The ordinary HTTP timeout bounds establishment and error-response handling,
  not successful stream lifetime. The caller context owns stream lifetime.
  An explicitly disabled/nonpositive timeout requires caller-owned cancellation.
- Drain `Frames` or call `Close()`; backpressure deliberately pauses reads.
  `Wait(ctx)` reports EOF, cancellation, invalid/oversize frames or read failure.
  Cancelling a **wait** does not close the subscription. Close and then wait to
  verify worker termination when abandoning a consumer.
- A wire frame is capped at `MaxSubscriptionFrameBytes` (8 MiB including newline).
- No automatic reconnect, snapshot classification or sparse-object merging is
  performed. Reconnect explicitly and re-establish trustworthy state. A live
  150-second observation established full-object arrays, not sparse-delta rules.
- Legacy `SubscribeClientWorkspaces` and other typed helpers remain value-only
  adapters. They cannot expose terminal errors, all field presence or empty-array
  boundaries; migrate stateful consumers to the new handle. Malformed frames now
  **terminate** legacy adapters too, rather than being logged and skipped; silent
  continuation after losing an event is unsafe. Legacy callers only see closure.

## Identity and explicit targets

`Login`, `LoginWithToken` and `SamlLogin` adopt a valid returned token and positive
user ID together. Authentication transitions serialize with cancellable waits.
Incomplete responses do not replace the selected actor. Token-store failure is
reported as `ErrTokenPersistence`; an already adopted actor is **not** rolled back
to a more privileged service identity. Zero expiry passed to `TokenStore` means
unknown expiry, not an expired token. A token-store **read failure during bootstrap**
makes the session unusable (`ErrTokenPersistence`) rather than falling back to a
service key. An empty store should return an empty token with no error. Recover
by constructing a new session after fixing the store, or with explicitly selected
authority; no automatic recovery changes the actor.

`GetCurrentUser` uses a known user ID or the observed `POST users/login` token
exchange (`token`, `remember:false`), caching the user ID without replacing the
selected credential or writing a replacement token to `TokenStore`. This is not
an HTTP read-only endpoint; one-time tokens must use explicit `LoginWithToken`
instead. It does not guess `/users/current`, which
returned 400 for authenticated requests in the disposable trial. Some SAML-issued
tokens do not support re-exchange: retain the live authenticated session/user ID
or explicitly reauthenticate; do not infer identity from workspace discovery.
SAML response adoption is fixture-tested, **not** a claim of live IdP validation.

Password helpers now use the observed wire fields: `SetUserPassword` sends
`new_password`; `ChangeUserPassword` sends `current_password` and `new_password`.
Server authorization still applies; no alternate-field retry is attempted.

Workspace selectors require a client ID and **exactly one** nonempty selector.
No selector no longer means workspace zero; multiple matches are errors. Supply
an explicit index (including zero) when that is the intended target. This also
applies to `canvus workspace open`: provide `--index`, `--name` or `--user`.
`Workspace.IndexPresent` and `IndexNull` distinguish missing/null from zero;
`GetWorkspace` rejects an absent or mismatched index rather than inventing one.

`ClientInfo.UserID` accepts numeric or string JSON and keeps `UserIDRaw` for the
wire representation. Client user association and `Workspace.User` (observed as
an email) are **not physical-operator identity**, canvas permission or native
control authority. Discovery visibility does not authorize a mutation.

## Geometry and camera

`WidgetBoundingBox` remains **raw stored model location/size**. It does not apply
scale, ancestry or Note registration. `Touches` now includes shared edges/corners,
matching Python/TypeScript. Shared valid-rectangle/raw-box fixtures run in all
three languages; that is not complete cross-language feature parity.

Use `WidgetCanvasBounds(id, snapshot, GeometryModel{...})` for rendered canvas
border-box bounds. Supply `RootWidgetID` identifying the actual `SharedCanvas`
widget, **not the canvas resource ID**, and include that identity-transform root
plus complete parent ancestry and positive uniform scales. Missing parents/types/transforms, cycles and invalid extents are
errors. For Notes explicitly choose `NotePadding`: 30 for the documented legacy
29px padding + 1px border registration, or 0 for normalized origins. Do not guess
the deployed model or rewrite raw locations. Legacy registration is
`location + 30 - 30*scale` at the root level, so the offset correctly cancels at
scale 1; the models differ at nonunit scale. Only unrotated uniform transforms
are supported. Native text fitting changes Note sizes asynchronously; a successful
write echo is not a final rendered-size observation.

`Workspace.ViewRectangle` is the established **scaled wire representation**, not
a visible canvas region. Reused equations from CanvusConsole and canvus-stress:

- Decode: `s = raw.width / workspace.width`, origin `-raw.position/s`,
  visible size `workspace.size/s` (`VisibleCanvasRegion`).
- Encode: fit/centre the canvas region, `s=min(wsW/regionW,wsH/regionH)`,
  raw position `-visibleOrigin*s`, raw size `workspace.size*s`
  (`ViewRectangleForRegion`). Larger `s` renders larger widgets.

`FrameWorkspaceRegion` and `SetWorkspaceViewport` require the intended canvas ID
and current open workspace metadata. `SetWorkspaceViewport` coordinates are now
explicitly **canvas pixels**, not raw wire values. For a raw PATCH use
`UpdateWorkspace`. Widget framing accepts the same explicit `NotePadding` model.

The helpers apply zero-position zoom, wait at least 100 ms, observe matching zoom
(up to one second), recheck the bound target, then pan. A combined zoom/pan PATCH
is not equivalent on affected clients. A clamped or unobserved zoom that cannot
be confirmed within the bound returns a non-replayable partial error: the helper
does not pan using an assumed scale or claim success. The initial actor and resolved index are
pinned; canvas/server/native user/state/size are rechecked. This is best-effort
race detection, **not** server compare-and-swap or a physical-operator lock.
Camera formulas target the documented scaled contract; do not assume other
native implementations or future contracts share it.

`OpenCanvasOnWorkspace` waits for matching canvas and `workspace_state:"open"`,
plus requested server/user when supplied. Merely seeing the canvas while loading
is not success. Point centring preserves the observed zoom; missing size/view
metadata returns an error instead of panicking. All waits honour context.

`CameraUpdateError` and `OpenCanvasError` preserve stage, known acknowledgment and
underlying HTTP/cancellation/accepted-response errors. A false acknowledgment
flag is **not proof of no mutation**. They are nonretryable operation outcomes;
reconcile instead of replaying or automatically undoing. Even a nil return does
not prove final rendering after the last accepted pan.

## Deliberate limits and compatibility

- Go only: Python/TypeScript authentication, streaming, target helpers and rendered
  geometry are not certified by this change. Their extras are preserved.
- Native non-owner camera requests returned HTTP 500 in the disposable trial.
  Preserve that status; do not relabel it as a verified 401/403 denial. Client-state
  transaction exceptions can map to generic 500 in native bridge source, but the
  exact deployed exception/server revision was not established.
- The historical audit's guessed current-user response, raw camera-centre and
  scale-only legacy-bounds hypotheses are superseded by the observed contracts
  and new boundary tests; do not claim every historical assertion is green.
- No live SAML authentication, new live camera mutation, deployment, release tag
  or product implementation is included. Use a verified revision when adopting;
  rollback means reverting the change and disabling affected capabilities, not
  restoring unsafe authority fallback or replay.

Offline checks: `go test -race -short ./...` from `go/sdk`, and
`python3 go/sdk/check_geometry_parity.py` from the repository root (Node 24+).
