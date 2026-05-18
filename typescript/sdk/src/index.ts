/**
 * @mt-canvus-tools/sdk — TypeScript SDK for the Canvus REST API.
 *
 * Public surface area is everything re-exported below. Anything not
 * re-exported here is internal and may change without a major bump.
 */

export { Session, createSession } from "./session.js";
export { Transport, safeJson } from "./transport.js";
export type { HttpMethod, RequestOptions, RequestBody } from "./transport.js";
export {
  loadConfig,
  buildConfig,
} from "./config.js";
export type { Config, SessionOptions } from "./config.js";
export { logger } from "./logging.js";
export type { Logger } from "./logging.js";
export { streamNdjson } from "./streaming.js";
export type { StreamOptions } from "./streaming.js";

// Errors
export {
  APIError,
  AuthError,
  CanvusError,
  NetworkError,
  NotFoundError,
  RateLimitError,
  ServerError,
  UnsupportedOperationError,
  ValidationError,
  isCanvusError,
  parseRetryAfter,
} from "./errors.js";
export type { CanvusErrorKind, ValidationIssue } from "./errors.js";

// Resource classes (re-exported so they can be referenced as types)
export { CanvasesResource } from "./resources/canvases.js";
export { WidgetsResource } from "./resources/widgets.js";
export type { CloneWidgetArgs, UploadMetadata } from "./resources/widgets.js";
export { AuthResource } from "./resources/auth.js";
export { UsersResource } from "./resources/users.js";
export { FoldersResource } from "./resources/folders.js";
export { AssetsResource } from "./resources/assets.js";
export type { MipmapOptions } from "./resources/assets.js";
export { ServerResource, flattenServerConfig } from "./resources/server.js";
export type { SetWorkspaceViewportOptions } from "./resources/server.js";

// All data-model types
export * from "./types/index.js";
