#!/usr/bin/env python3
"""Check shared Go/Python/TS geometry fixtures; no dependencies or live API.

Go consumes the same fixture in TestSharedGeometryParityFixtures. Run this
script with Python 3.10+ and Node 24+ (native TypeScript stripping).
"""
from dataclasses import asdict
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
from types import SimpleNamespace

root = Path(__file__).resolve().parents[2]
fixture_path = root / "go/sdk/canvus/testdata/geometry-parity.json"
fixture = json.loads(fixture_path.read_text())
source = root / "python/sdk/src/canvus_sdk/extras/geometry.py"
spec = importlib.util.spec_from_file_location("parity_geometry", source)
module = importlib.util.module_from_spec(spec)
sys.modules[spec.name] = module
spec.loader.exec_module(module)
for case in fixture["rectangles"]:
    a, b = module.Rectangle(**case["a"]), module.Rectangle(**case["b"])
    assert module.contains(a, b) == case["contains"], case["name"]
    assert module.touches(a, b) == case["touches"], case["name"]
assert asdict(module.widget_bounding_box(SimpleNamespace(**fixture["raw_widget"]))) == fixture["raw_bounds"]
print("Python: shared rectangle/edge/raw-bounds fixtures passed", flush=True)
subprocess.run([
    "node", "--input-type=module", "-e", """
import fs from 'node:fs';
import assert from 'node:assert/strict';
import {pathToFileURL} from 'node:url';
const g = await import(pathToFileURL(process.argv[1]));
const fixture = JSON.parse(fs.readFileSync(process.argv[2], 'utf8'));
for (const c of fixture.rectangles) {
  assert.equal(g.contains(c.a,c.b),c.contains,c.name);
  assert.equal(g.touches(c.a,c.b),c.touches,c.name);
}
assert.deepEqual(g.widgetBoundingBox(fixture.raw_widget),fixture.raw_bounds);
console.log('TypeScript: shared rectangle/edge/raw-bounds fixtures passed');
""", str(root / "typescript/sdk/src/extras/geometry.ts"), str(fixture_path)
], check=True)
