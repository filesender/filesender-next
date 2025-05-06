import { test, expect } from '@playwright/test';
import { uploadTestFile } from './fixtures/upload-helper';

test('should download', async ({ page }) => {
    const { cleanup } = await uploadTestFile(page);
    try {
        const buttons = await page.locator('button').all();

        let downloadTriggered = false;

        page.on('request', request => {
            if (request.url().includes('/download') && !request.url().includes('/api')) {
                downloadTriggered = true;
            }
        });

        for (const button of buttons) {
            if ((await button.innerText()).trim() === "Download") {
                await button.click();
                break;
            }
        }

        await page.waitForTimeout(2000);
        expect(downloadTriggered).toBe(true);

    } finally {
        cleanup();
    }
});
