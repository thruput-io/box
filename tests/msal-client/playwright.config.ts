import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 60000,
  expect: {
    timeout: 10000
  },
  use: {
    baseURL: 'https://msal-client.web.internal/',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    actionTimeout: 1000,
    navigationTimeout: 1000,
  },
  webServer: {
    command: 'npm run preview -- --port 8080 --host',
    port: 8080,
    reuseExistingServer: !process.env.CI,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
