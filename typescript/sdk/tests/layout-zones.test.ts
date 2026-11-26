import { afterEach, expect, it, vi } from "vitest";
import { createSession } from "../src/index.js";

afterEach(() => vi.unstubAllGlobals());
it("routes layout-zone CRUD and preserves array payloads and generic fields", async () => {
  const zone = { id: "z", widget_type: "LayoutZone", layout_mode: "grid", slot_order: [], main_promoted: false, gap_fraction: .015, layout_version: 1 };
  const fetcher = vi.fn(async (url: string, options?: RequestInit) => {
    expect(String(url)).toContain("canvases/c/");
    const list = String(url).endsWith("layout-zones") && options?.method === "GET";
    return new Response(options?.method === "DELETE" ? null : JSON.stringify(list ? [zone] : zone), {status: options?.method === "DELETE" ? 204 : 200, headers: {"content-type":"application/json"}});
  });
  vi.stubGlobal("fetch", fetcher);
  const session = createSession({ baseUrl: "https://example.invalid/api/v1/" });
  expect((await session.widgets.layoutZones.create("c", {})).slot_order).toEqual([]);
  await session.widgets.updateAny("c", "z", {widget_type:"LayoutZone",layout_mode:"grid",slot_order:[]});
  expect(JSON.parse(String(fetcher.mock.calls[1]?.[1]?.body))).toEqual({widget_type:"LayoutZone",layout_mode:"grid",slot_order:[]});
  expect(String(fetcher.mock.calls[1]?.[0])).toContain("layout-zones/z");
  expect((await session.widgets.layoutZones.get("c", "z")).layout_version).toBe(1);
  expect(await session.widgets.layoutZones.list("c")).toHaveLength(1);
  await session.widgets.layoutZones.delete("c", "z");
  expect(fetcher.mock.calls.map(call => call[1]?.method)).toEqual(["POST","PATCH","GET","GET","DELETE"]);
});
