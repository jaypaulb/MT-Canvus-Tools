/**
 * User identity + color-assignment routes.
 *
 * Ports the legacy `/identify-user`, `/list-users`, `/delete-users`
 * endpoints. Color storage lives in `users.json` (demo-grade — see
 * `lib/store.ts`).
 */

import { Hono } from "hono";
import { zValidator } from "@hono/zod-validator";
import { z } from "zod";
import type { AppEnv } from "../app-env.js";
import {
  generateColorVariation,
  getTeamBaseColor,
  isValidTeam,
  type TeamNumber,
} from "../lib/colors.js";

const identifyBody = z.object({
  team: z.number().int().min(1).max(7),
  name: z.string().trim().min(1),
});

const deleteBody = z.object({
  all: z.boolean().optional(),
  users: z
    .array(
      z.object({
        team: z.number().int().min(1).max(7),
        name: z.string().min(1),
      }),
    )
    .optional(),
});

export function usersRoutes(): Hono<AppEnv> {
  const app = new Hono<AppEnv>();

  app.post("/identify-user", zValidator("json", identifyBody), async (c) => {
    const { team, name } = c.req.valid("json");
    if (!isValidTeam(team)) {
      return c.json({ success: false, error: "Invalid team selected." }, 400);
    }
    const store = c.var.userStore;
    const users = await store.read();
    const key = team.toString();
    if (!users[key]) users[key] = {};
    const existing = users[key][name];
    if (existing !== undefined) {
      return c.json({ success: true, color: existing });
    }
    const newColor = generateColorVariation(getTeamBaseColor(team as TeamNumber));
    users[key][name] = newColor;
    await store.write(users);
    return c.json({ success: true, color: newColor });
  });

  app.get("/list-users", async (c) => {
    const users = await c.var.userStore.read();
    const list: Array<{ team: number; name: string; color: string }> = [];
    for (const [team, members] of Object.entries(users)) {
      for (const [name, color] of Object.entries(members)) {
        list.push({ team: Number.parseInt(team, 10), name, color });
      }
    }
    return c.json({ success: true, users: list });
  });

  app.post("/delete-users", zValidator("json", deleteBody), async (c) => {
    const body = c.req.valid("json");
    const store = c.var.userStore;
    if (body.all === true) {
      await store.write({});
      return c.json({ success: true, message: "All users deleted." });
    }
    if (!body.users || body.users.length === 0) {
      return c.json({ success: false, error: "No users specified for deletion." }, 400);
    }
    const current = await store.read();
    let deleted = 0;
    for (const u of body.users) {
      const key = u.team.toString();
      const bucket = current[key];
      if (bucket && bucket[u.name] !== undefined) {
        delete bucket[u.name];
        deleted++;
        if (Object.keys(bucket).length === 0) delete current[key];
      }
    }
    await store.write(current);
    return c.json({ success: true, message: `${deleted.toString()} users deleted.` });
  });

  return app;
}
