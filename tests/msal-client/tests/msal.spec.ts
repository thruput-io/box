import { test, expect } from '@playwright/test';

test.describe('MSAL Login Flow', () => {
  test('should login successfully with state preservation', async ({ page }) => {
    // Console log listener
    page.on('console', msg => console.log('BROWSER LOG:', msg.text()));
    page.on('pageerror', err => console.log('BROWSER ERROR:', err.message));

    // Navigate to the SPA
    await page.goto('https://msal-client.web.internal/');

    // Click login
    await page.click('#login');

    // Wait for redirect to Identity Mock
    await expect(page).toHaveURL(/.*login\.microsoftonline\.com.*/);

    // Verify state is in the URL
    const url = new URL(page.url());
    const state = url.searchParams.get('state');
    expect(state).toContain('custom-state-value-123');

    // Perform Quick Login (Debug feature we added to the mock)
    // Find the login button for 'diego' in the debug list
    await page.locator('.user-item', { hasText: 'diego' }).locator('button').click();

    // Wait for redirect back to the SPA
    await expect(page).toHaveURL(/.*msal-client\.web\.internal.*/);

    // Verify account info is displayed
    const accountInfo = page.locator('#account-info');
    await expect(accountInfo).toContainText('user@abroad.com');
    await expect(accountInfo).toContainText('Diego Admin');

    // Verify state info is displayed (handled by our SPA logic)
    const stateInfo = page.locator('#state-info');
    await expect(stateInfo).toContainText('State received: custom-state-value-123');

    // Verify tokens were acquired (roles should be present)
    const tokenInfo = page.locator('#token-info');
    await expect(tokenInfo).toContainText('idToken');
    await expect(tokenInfo).toContainText('accessToken');

    // Test Silent Token acquisition
    console.log("Testing Silent Token acquisition...");
    await page.click('#silent-token');
    
    // Verify it still has tokens (and possibly updated)
    await expect(tokenInfo).toContainText('accessToken');
    // Ensure we are still on the same page (no redirect)
    await expect(page).toHaveURL(/.*msal-client\.web\.internal.*/);
    console.log("Silent Token acquisition successful");
  });
});
