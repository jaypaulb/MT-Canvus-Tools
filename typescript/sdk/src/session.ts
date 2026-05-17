import { buildConfig, type Config, type SessionOptions } from "./config.js";
import { Transport } from "./transport.js";
import { CanvasesResource } from "./resources/canvases.js";
import { WidgetsResource } from "./resources/widgets.js";
import { AuthResource } from "./resources/auth.js";
import { UsersResource } from "./resources/users.js";
import { FoldersResource } from "./resources/folders.js";
import { AssetsResource } from "./resources/assets.js";
import { ServerResource } from "./resources/server.js";

/**
 * Top-level entry point of the Canvus TypeScript SDK.
 *
 * A `Session` bundles the runtime config, the HTTP transport, and one
 * instance of every resource namespace. Create one session per
 * application; resources are cheap to retain.
 *
 * @example
 * ```ts
 * import { createSession } from "@mt-canvus-tools/sdk";
 *
 * const session = createSession({
 *   baseUrl: "https://canvus.example.com/api/v1/",
 *   apiKey: process.env.CANVUS_API_KEY,
 * });
 *
 * for (const canvas of await session.canvases.list()) {
 *   console.log(canvas["canvas-name"]);
 * }
 * ```
 */
export class Session {
  /** Frozen runtime configuration used by every resource. */
  public readonly config: Config;
  /** Low-level HTTP transport. Useful for endpoints not yet on the SDK. */
  public readonly transport: Transport;

  public readonly canvases: CanvasesResource;
  public readonly widgets: WidgetsResource;
  public readonly auth: AuthResource;
  public readonly users: UsersResource;
  public readonly folders: FoldersResource;
  public readonly assets: AssetsResource;
  public readonly server: ServerResource;

  /**
   * Construct a session directly from a parsed {@link Config} (e.g. from
   * `loadConfig()`). Most callers should use {@link createSession}.
   */
  constructor(config: Config) {
    this.config = config;
    this.transport = new Transport(config);
    this.canvases = new CanvasesResource(this.transport);
    this.widgets = new WidgetsResource(this.transport);
    this.auth = new AuthResource(this.transport);
    this.users = new UsersResource(this.transport);
    this.folders = new FoldersResource(this.transport);
    this.assets = new AssetsResource(this.transport);
    this.server = new ServerResource(this.transport);
  }
}

/**
 * Create a new {@link Session} from explicit options.
 *
 * @throws {ValidationError} If `baseUrl` is invalid or any numeric field is out of range.
 */
export function createSession(opts: SessionOptions): Session {
  return new Session(buildConfig(opts));
}
