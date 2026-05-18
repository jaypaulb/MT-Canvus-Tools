import { beforeEach, describe, expect, it, vi } from "vitest";
import { logger } from "../../src/logging.js";
import {
  disableApiWarnings,
  enableApiWarnings,
  KnownWarnings,
  resetWarnings,
  WarningNoteTitleNotExposed,
  WarningTableGridSizeImmutable,
  warnAlways,
  warnOnce,
} from "../../src/extras/warnings.js";

describe("APIWarning registry", () => {
  beforeEach(() => {
    resetWarnings();
    enableApiWarnings();
  });

  it("known warnings catalogue contains every Go warning code", () => {
    expect(KnownWarnings[WarningNoteTitleNotExposed.code]).toBeDefined();
    expect(KnownWarnings[WarningTableGridSizeImmutable.code]).toBeDefined();
  });

  it("warnOnce emits exactly once per code", () => {
    const warn = vi.spyOn(logger, "warn").mockImplementation((..._args) => undefined);
    warnOnce(WarningNoteTitleNotExposed);
    warnOnce(WarningNoteTitleNotExposed);
    expect(warn).toHaveBeenCalledTimes(1);
    warn.mockRestore();
  });

  it("disableApiWarnings suppresses both helpers", () => {
    const warn = vi.spyOn(logger, "warn").mockImplementation((..._args) => undefined);
    disableApiWarnings();
    warnOnce(WarningNoteTitleNotExposed);
    warnAlways(WarningNoteTitleNotExposed);
    expect(warn).not.toHaveBeenCalled();
    warn.mockRestore();
  });
});
