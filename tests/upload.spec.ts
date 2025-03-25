import { test, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';
import os from 'os';
import { randomUUID } from 'crypto';

test('should upload a file', async ({ page }) => {
    // Create file in the system's temp directory
    const fileName = `temp-file-${randomUUID()}.txt`;
    const tempDir = os.tmpdir();
    const filePath = path.join(tempDir, fileName);
    fs.writeFileSync(filePath, 'This is a test file content.'); // 28 bytes

    try {
        await page.goto('http://localhost:8080/');

        const fileChooserPromise = page.waitForEvent('filechooser');
        await page.click('#files-selector');
        const fileChooser = await fileChooserPromise;

        await fileChooser.setFiles(filePath);

        const uploadedFileName = await page.$eval('#files-selector', (input) => {
            return (input as HTMLInputElement).files?.[0]?.name;
        });

        expect(uploadedFileName).toBe(fileName);

        await page.click('input[type="submit"]');
        await page.waitForURL('**/upload/**', { timeout: 5000 });

        const pageBody = await page.content();
        expect(pageBody).toContain("28 bytes");
        expect(pageBody).toContain("1 file");
    } finally {
        // Always cleanup, even if test fails
        if (fs.existsSync(filePath)) {
            fs.unlinkSync(filePath);
        }
    }
});
