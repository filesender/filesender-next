import { test, expect } from '@playwright/test';

test('should return a successful response', async ({ page }) => {
    const response = await page.goto('http://localhost:8080');

    expect(response?.status()).toBe(200);
    await expect(page).toHaveTitle("Upload");
});

test('should contain a file input', async ({ page }) => {
    await page.goto('http://localhost:8080');

    const fileInputs = page.locator('input[type="file"]');
    await expect(fileInputs).toHaveCount(1);
});
