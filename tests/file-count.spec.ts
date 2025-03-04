import { test, expect } from '@playwright/test';

test('should return a successful response', async ({ page }) => {
    const response = await page.goto('http://localhost:8080/file-count');
    expect(response?.status()).toBe(200);
    expect(await response?.text()).toContain("There are 0 files uploaded");
});
