import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: ["src/**/*.test.ts", "tests/**/*.test.ts"],
    exclude: ["tests/integration/**", "node_modules", "dist"],
    // Phase 4b: use forks/singleFork so the esbuild service shared between
    // worker pool threads doesn't OOM under transform pressure. Tests are
    // mock-only and complete in <15s; serialised execution is fine here.
    pool: "forks",
    poolOptions: { forks: { singleFork: true } },
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
      // Phase 4d Round C: thresholds reclaimed after auth/users/server unit
      // coverage was added (those resource files now hit ~100% statements).
      // Values sit ~5pp below the actual numbers to leave a small drift
      // margin without permitting silent regressions.
      thresholds: { lines: 65, branches: 70, functions: 60, statements: 65 },
      include: ["src/**/*.ts"],
      exclude: ["src/**/*.test.ts", "src/index.ts", "src/extras/index.ts"],
    },
  },
});
