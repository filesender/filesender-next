import { test, expect } from '@playwright/test';
import { uploadTestFile } from './fixtures/upload-helper';

test('should download', async ({ page }) => {
    const { cleanup } = await uploadTestFile(page);
    try {
        const buttons = await page.locator('button').all();

        // Set up a download watcher
        const [download] = await Promise.all([
            page.waitForEvent('download', req => 
                req.url().includes('/download')
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
        expect(download.url()).toMatch(/\/download\/.+/);

        await download.path(); // Waits until download is complete
        const suggestedFilename = download.suggestedFilename();
        expect(suggestedFilename).toMatch(/^temp-file-.+\.txt$/);

    } finally {
        cleanup();
    }
});
