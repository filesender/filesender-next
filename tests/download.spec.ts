import { test, expect } from '@playwright/test';
import { uploadTestFile } from './fixtures/upload-helper';

test('should upload a file', async ({ page }) => {
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
