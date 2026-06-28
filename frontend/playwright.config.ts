import { defineConfig, devices } from "@playwright/test"

// E2E config for shrt. Tests assume the full local stack is running:
//   - Go backend on :8080 (make dev / go run ./cmd/shrt)
//   - Next.js frontend on :3000 (npm run dev)
//   - Postgres + Redis (make docker-up)
//
// Run with: npm run e2e  (add --ui for the inspector).
//
// FRONTEND_URL / BACKEND_URL env vars override the defaults so the same suite
// can run against a deployed environment.
const FRONTEND_URL = process.env.E2E_FRONTEND_URL ?? "http://localhost:3000"

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? "github" : "list",
  use: {
    baseURL: FRONTEND_URL,
    trace: "on-first-retry",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
})
