import { test, expect } from '@playwright/test';

test('should return a not found response', async ({ page }) => {
    const response = await page.goto('http://localhost:8080/abababab');
    expect(response?.status()).toBe(404);
});
