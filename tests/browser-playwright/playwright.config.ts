import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 10_000,
  retries: 1,
  workers: 1,
  use: {
    baseURL: 'https://browser.web.internal',
    ignoreHTTPSErrors: true,
    headless: true,
    trace: 'retain-on-failure'
  }
});
