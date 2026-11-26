import json

import pytest
import respx

from canvus_sdk import Client
from canvus_sdk.models import LayoutZone


@pytest.mark.asyncio
async def test_layout_zone_transport(client: Client) -> None:
    zone = LayoutZone(id="z").model_dump()
    with respx.mock(base_url=client.base_url, assert_all_called=True) as mock:
        create = mock.post("canvases/c/layout-zones").respond(200, json=zone)
        patch = mock.patch("canvases/c/layout-zones/z").respond(200, json=zone)
        mock.get("canvases/c/layout-zones/z").respond(200, json=zone)
        mock.get("canvases/c/layout-zones").respond(200, json=[zone])
        mock.get("canvases/c/widgets").respond(200, json=[zone])
        mock.delete("canvases/c/layout-zones/z").respond(204)
        created = await client.widgets.layout_zones.create("c", {})
        assert created.slot_order == []
        assert created.size == {"width": 960, "height": 540}
        assert json.loads(create.calls[0].request.content) == {}
        await client.widgets.update_any("c", "z", {"widget_type": "LayoutZone", "layout_mode": "grid", "slot_order": []})
        assert json.loads(patch.calls[0].request.content) == {"widget_type": "LayoutZone", "layout_mode": "grid", "slot_order": []}
        assert (await client.widgets.layout_zones.get("c", "z")).layout_version == 1
        assert len(await client.widgets.layout_zones.list("c")) == 1
        assert (await client.widgets.list("c"))[0].model_dump()["slot_order"] == []
        await client.widgets.layout_zones.delete("c", "z")
