// Phase 4b §4.3 #21: APIWarning registry.
//
// Mirrors Go's `warnings.go:9-135`. Catalogues known Canvus API limitations
// so SDK consumers can surface them once per process via the shared `logger`.

import { logger } from "../logging.js";

/** A single known limitation in the Canvus API. */
export interface APIWarning {
  readonly code: string;
  readonly description: string;
  readonly workaround: string;
  readonly issueUrl: string;
}

/** Note widget `title` is silently ignored on PATCH/POST. */
export const WarningNoteTitleNotExposed: APIWarning = Object.freeze({
  code: "NOTE_TITLE_NOT_EXPOSED",
  description:
    "Note widget 'title' field is not exposed by the Canvus API. Title values in requests are ignored and responses will not include the title.",
  workaround: "Use the 'name' field instead for identifying notes.",
  issueUrl: "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/38",
});

/** VideoInput widget `title` field is not API-visible. */
export const WarningVideoInputTitleNotExposed: APIWarning = Object.freeze({
  code: "VIDEOINPUT_TITLE_NOT_EXPOSED",
  description: "VideoInput widget 'title' field is not exposed by the Canvus API.",
  workaround: "No workaround available. Await Canvus API fix.",
  issueUrl: "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/13",
});

/** PDF resize updates bounding box but not actual content size. */
export const WarningPdfSizeBug: APIWarning = Object.freeze({
  code: "PDF_SIZE_BUG",
  description:
    "PDF widget size changes via PATCH update the bounding box but the actual PDF content stays at its original size.",
  workaround: "Avoid resizing PDFs via API. Delete and recreate if different size is needed.",
  issueUrl: "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/15",
});

/** Image resize via PATCH does not preserve aspect ratio. */
export const WarningImageAspectRatioNotPreserved: APIWarning = Object.freeze({
  code: "IMAGE_ASPECT_RATIO_NOT_PRESERVED",
  description: "Image widget size changes via PATCH do not preserve aspect ratio.",
  workaround:
    "Calculate correct aspect-ratio-preserving dimensions before calling widgets.images.update.",
  issueUrl: "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/39",
});

/** Video resize via PATCH does not preserve aspect ratio. */
export const WarningVideoAspectRatioNotPreserved: APIWarning = Object.freeze({
  code: "VIDEO_ASPECT_RATIO_NOT_PRESERVED",
  description: "Video widget size changes via PATCH do not preserve aspect ratio.",
  workaround:
    "Calculate correct aspect-ratio-preserving dimensions before calling widgets.videos.update.",
  issueUrl: "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/39",
});

/** Table `grid_size` is silently ignored on PATCH. */
export const WarningTableGridSizeImmutable: APIWarning = Object.freeze({
  code: "TABLE_GRID_SIZE_IMMUTABLE",
  description: "Table widget 'grid_size' is set at creation time and silently ignored on PATCH.",
  workaround: "Omit grid_size from PATCH requests; recreate the table to change dimensions.",
  issueUrl: "https://gitlab.multitaction.com/swrd/conan/canvus/canvus-app/-/issues/-",
});

/** The full set of known warnings, keyed by `code`. */
export const KnownWarnings: Readonly<Record<string, APIWarning>> = Object.freeze({
  [WarningNoteTitleNotExposed.code]: WarningNoteTitleNotExposed,
  [WarningVideoInputTitleNotExposed.code]: WarningVideoInputTitleNotExposed,
  [WarningPdfSizeBug.code]: WarningPdfSizeBug,
  [WarningImageAspectRatioNotPreserved.code]: WarningImageAspectRatioNotPreserved,
  [WarningVideoAspectRatioNotPreserved.code]: WarningVideoAspectRatioNotPreserved,
  [WarningTableGridSizeImmutable.code]: WarningTableGridSizeImmutable,
});

let warningsEnabled = process.env.CANVUS_SDK_DISABLE_WARNINGS !== "1";
const issued = new Set<string>();

/** Suppress all subsequent warnings emitted by {@link warnOnce}/{@link warnAlways}. */
export function disableApiWarnings(): void {
  warningsEnabled = false;
}

/** Re-enable warning output. */
export function enableApiWarnings(): void {
  warningsEnabled = true;
}

/** Forget which warnings have already been emitted (primarily for tests). */
export function resetWarnings(): void {
  issued.clear();
}

/** Emit `warning` at most once per process lifetime. */
export function warnOnce(warning: APIWarning): void {
  if (!warningsEnabled) return;
  if (issued.has(warning.code)) return;
  issued.add(warning.code);
  logger.warn(
    {
      code: warning.code,
      description: warning.description,
      workaround: warning.workaround,
      issue: warning.issueUrl,
    },
    "canvus SDK warning",
  );
}

/** Emit `warning` on every call (use sparingly). */
export function warnAlways(warning: APIWarning): void {
  if (!warningsEnabled) return;
  logger.warn(
    { code: warning.code, description: warning.description },
    "canvus SDK warning",
  );
}
