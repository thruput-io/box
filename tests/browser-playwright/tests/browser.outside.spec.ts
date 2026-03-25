import { expect, test } from '@playwright/test';

test('browser UI is reachable from outside and renders Firefox desktop', async ({ page }) => {
  const response = await page.goto('/', { waitUntil: 'domcontentloaded' });
  expect(response).not.toBeNull();
  expect(response!.status()).toBe(200);

  await expect(page.locator('#noVNC_container')).toBeVisible();
});
