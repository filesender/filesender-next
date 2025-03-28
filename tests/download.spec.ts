import { test, expect } from '@playwright/test';
import { uploadTestFile } from './fixtures/upload-helper';

test('should download link be available', async ({ page }) => {
    const { fileName, cleanup } = await uploadTestFile(page);
    try {
        const uploadPageBody = await page.content();
        const downloadLink = uploadPageBody.split("Download: ")[1].split("<")[0];

        await page.goto(downloadLink);
        const transferPageBody = await page.content();
        expect(transferPageBody).toContain("28 bytes");
        expect(transferPageBody).toContain("1 file");
        expect(transferPageBody).toContain(fileName);

    } finally {
        cleanup();
    }
});

test('should download', async ({ page, context }) => {
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
                    if ((await button.innerText()).trim() === "Download All") {
                        await button.click();
                        break;
                    }
                }
            })(),
        ]);

        console.log('Download request URL:', download.url());
        expect(download.url()).toMatch(/\/api\/v1\/download\/.+\/.+\/all/);

        await download.path(); // Waits until download is complete
        const suggestedFilename = download.suggestedFilename();
        expect(suggestedFilename).toMatch(/\.zip$/i);

    } finally {
        cleanup();
    }
});
