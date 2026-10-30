import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  outputDir: "../.ui-check/results",
  reporter: "line",
  webServer: {
    // `.next-dev` memisahkan hasil build dev dari `frontend/.next` yang juga
    // ditempel ke image Docker; tanpa ini `next dev` dapat menimpa CSS/halaman
    // container :3001 saat dijalankan bersamaan.
    command: "npm run dev -- --hostname 127.0.0.1 --port 3002",
    env: { NEXT_DIST_DIR: ".next-dev" },
    url: "http://127.0.0.1:3002",
    reuseExistingServer: true,
  },
  use: {
    baseURL: "http://127.0.0.1:3002",
    channel: "chrome",
    screenshot: "only-on-failure",
  },
  projects: [
    { name: "desktop", use: { ...devices["Desktop Chrome"] } },
    { name: "mobile", use: { ...devices["Pixel 7"] } },
  ],
});
