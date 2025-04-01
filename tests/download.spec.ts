import { test, expect } from '@playwright/test';
import { uploadTestFile } from './fixtures/upload-helper';

test('should download link be available', async ({ page }) => {
    const { cleanup } = await uploadTestFile(page);
    try {
        const uploadPageBody = await page.content();
        const downloadLink = uploadPageBody.split("Download: ")[1].split("<")[0];

        await page.goto(downloadLink);
        const transferPageBody = await page.content();
        expect(transferPageBody).toContain("2048 bytes");

    } finally {
        cleanup();
    }
});

test('should download', async ({ page }) => {
    const { cleanup } = await uploadTestFile(page);
    try {
        const uploadPageBody = await page.content();
        const downloadLink = uploadPageBody.split("Download: ")[1].split("<")[0];

        await page.goto(downloadLink);

        const buttons = await page.locator('button').all();

        // Set up a download watcher
        const [download] = await Promise.all([
            page.waitForEvent('download', req => 
                req.url().includes('/api/v1/download')
            ),
            (async () => {
                for (const button of buttons) {
                    if ((await button.innerText()).trim() === "Download") {
                        await button.click();
                        break;
                    }
                }
            })(),
        ]);

        console.log('Download request URL:', download.url());
        expect(download.url()).toMatch(/\/api\/v1\/download\/.+\/.+/);

        await download.path(); // Waits until download is complete
        const suggestedFilename = download.suggestedFilename();
        expect(suggestedFilename).toMatch(/\.tar$/i);

    } finally {
        cleanup();
    }
});
