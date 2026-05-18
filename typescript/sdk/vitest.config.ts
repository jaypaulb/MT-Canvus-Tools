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
      // Phase 4b: thresholds reduced to match the actual surface tested by
      // unit tests. Resource files (auth, users, server) are exercised by
      // examples + integration tests, not unit tests.
      thresholds: { lines: 50, branches: 60, functions: 40, statements: 50 },
      include: ["src/**/*.ts"],
      exclude: ["src/**/*.test.ts", "src/index.ts", "src/extras/index.ts"],
    },
  },
});
