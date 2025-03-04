import { test, expect } from '@playwright/test';

test('should return a successful response', async ({ page }) => {
    const response = await page.goto('http://localhost:8080');
    expect(response?.status()).toBe(200);
    expect(await response?.text()).toContain("Hello, world!");
});
