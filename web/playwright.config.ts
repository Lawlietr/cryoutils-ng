import { defineConfig, devices } from '@playwright/test'

const baseUrl = process.env.BASE_URL ?? 'http://127.0.0.1:33031'

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: 0,
  workers: 1,
  reporter: 'list',
  use: {
    baseURL: baseUrl,
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'Steam Deck',
      use: { ...devices['Steam Deck'] },
    },
  ],
  webServer: process.env.CI ? {
    command: 'npm run preview',
    url: 'http://127.0.0.1:4173',
    reuseExistingServer: true,
    timeout: 10000,
  } : undefined,
})
