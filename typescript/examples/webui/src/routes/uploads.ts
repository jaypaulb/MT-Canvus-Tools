/**
 * Upload routes — port of legacy `/create-note`, `/upload-item`,
 * `/list-target-notes`, `/delete-target-notes`.
 *
 * - Notes are created via JSON POST through `session.widgets.notes.create`.
 * - Images / videos / PDFs are uploaded via the SDK's per-type
 *   `upload(canvasId, blob, filename, meta)` helpers (multipart with
 *   `data` + `json` parts — same shape the legacy server emitted).
 *
 * The legacy "place near Team_N_Target" behaviour is preserved: the
 * upload payload's location is offset randomly within ±300 of the team
 * target widget's position.
 */

import { Hono } from "hono";
import { zValidator } from "@hono/zod-validator";
import { z } from "zod";
import type { AppEnv } from "../app-env.js";
import type { Widget, Note } from "@mt-canvus-tools/sdk";
import { getTeamBaseColor, isValidTeam, type TeamNumber } from "../lib/colors.js";

const EXT_TO_KIND: Record<string, "image" | "video" | "pdf"> = {
  jpg: "image",
  jpeg: "image",
  png: "image",
  gif: "image",
  bmp: "image",
  tiff: "image",
  mp4: "video",
  avi: "video",
  mov: "video",
  wmv: "video",
  mkv: "video",
  pdf: "pdf",
};

const createNoteBody = z.object({
  team: z.number().int().min(1).max(7),
  name: z.string().trim().min(1),
  text: z.string().trim().min(1),
  // #RRGGBBAA hex
  color: z.string().regex(/^#[0-9A-Fa-f]{8}$/),
});

const deleteTargetNotesBody = z.object({
  noteIds: z.array(z.string().min(1)).min(1),
});

function nowStamp(): string {
  // matches legacy "DD MMM HH:MM:SS" without commas
  return new Date()
    .toLocaleString("en-GB", {
      day: "2-digit",
      month: "short",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    })
    .replace(",", "");
}

function randomOffset(): number {
  return Math.floor(Math.random() * 601) - 300;
}

/** Find the `Team_N_Target` note for placing uploads near it. */
async function findTeamTarget(
  widgets: readonly Widget[],
  team: TeamNumber,
): Promise<{ x: number; y: number } | undefined> {
  const title = `Team_${team.toString()}_Target`;
  const found = widgets.find((w) => (w as unknown as { title?: string }).title === title) as
    | (Widget & { location?: { x: number; y: number } })
    | undefined;
  return found?.location;
}

export function uploadsRoutes(): Hono<AppEnv> {
  const app = new Hono<AppEnv>();

  app.post("/create-note", zValidator("json", createNoteBody), async (c) => {
    const { team, name, text, color } = c.req.valid("json");
    if (!isValidTeam(team)) {
      return c.json({ success: false, error: "Invalid team." }, 400);
    }
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    const widgets = await session.widgets.list(canvasId);
    const targetLoc = await findTeamTarget(widgets, team as TeamNumber);
    const baseX = targetLoc?.x ?? 0;
    const baseY = targetLoc?.y ?? 0;

    const created = await session.widgets.notes.create(canvasId, {
      auto_text_color: true,
      background_color: color,
      depth: 100 + Math.floor(Math.random() * 201),
      location: { x: baseX + randomOffset(), y: baseY + randomOffset() },
      pinned: false,
      scale: 1,
      size: { width: 300, height: 300 },
      state: "normal",
      text,
      title: `${name} @ (${nowStamp()})`,
    } as Parameters<typeof session.widgets.notes.create>[1]);
    return c.json({ success: true, id: created.id });
  });

  // Hono multipart upload — accepts file under field name "file" and
  // form fields "team", "name".
  app.post("/upload-item", async (c) => {
    const form = await c.req.parseBody();
    const file = form["file"];
    if (!(file instanceof File)) {
      return c.json({ success: false, error: "No file uploaded." }, 400);
    }
    const teamRaw = form["team"];
    const nameRaw = form["name"];
    const team = typeof teamRaw === "string" ? Number.parseInt(teamRaw, 10) : NaN;
    const name = typeof nameRaw === "string" ? nameRaw : "";
    if (!isValidTeam(team) || name.length === 0) {
      return c.json({ success: false, error: "Team or name missing in upload request." }, 400);
    }

    const ext = (file.name.split(".").pop() ?? "").toLowerCase();
    const kind = EXT_TO_KIND[ext];
    if (!kind) {
      return c.json(
        { success: false, error: `File type "${ext}" is not supported.` },
        400,
      );
    }

    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    const widgets = await session.widgets.list(canvasId);
    const targetLoc = await findTeamTarget(widgets, team as TeamNumber);
    const x = (targetLoc?.x ?? 0) + randomOffset();
    const y = (targetLoc?.y ?? 0) + randomOffset();

    const blob = await file.arrayBuffer().then((buf) => new Blob([buf]));
    const meta = {
      title: `${name} uploaded ${file.name} @ (${nowStamp()})`,
      location: { x, y },
      pinned: false,
      scale: 1,
      depth: 0,
    } as const;

    const helper =
      kind === "image"
        ? session.widgets.images.upload.bind(session.widgets.images)
        : kind === "video"
          ? session.widgets.videos.upload.bind(session.widgets.videos)
          : session.widgets.pdfs.upload.bind(session.widgets.pdfs);

    const created = await helper(canvasId, blob, file.name, meta);
    return c.json({
      success: true,
      message: `File "${file.name}" uploaded.`,
      id: (created as unknown as { id: string }).id,
    });
  });

  app.get("/list-target-notes", async (c) => {
    const widgets = await c.var.session.widgets.list(c.var.mutableConfig.canvasId);
    const notes: Note[] = [];
    for (const w of widgets) {
      const title = (w as unknown as { title?: string }).title;
      if (w.widget_type === "Note" && title && /^Team_\d+_Target$/.test(title)) {
        notes.push(w as Note);
      }
    }
    return c.json({ success: true, notes });
  });

  app.post("/delete-target-notes", zValidator("json", deleteTargetNotesBody), async (c) => {
    const { noteIds } = c.req.valid("json");
    const session = c.var.session;
    const canvasId = c.var.mutableConfig.canvasId;
    let deleted = 0;
    let failed = 0;
    for (const id of noteIds) {
      try {
        await session.widgets.notes.delete(canvasId, id);
        deleted++;
      } catch (err) {
        failed++;
        c.var.logger.warn({ id, err: (err as Error).message }, "delete-target-notes: failed");
      }
    }
    return c.json({
      success: true,
      message: `${deleted.toString()} notes deleted, ${failed.toString()} failed.`,
    });
  });

  return app;
}
