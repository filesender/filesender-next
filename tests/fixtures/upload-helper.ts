import fs from 'fs';
import path from 'path';
import os from 'os';
import { randomUUID } from 'crypto';
import { Page } from '@playwright/test';

export async function uploadTestFile(page: Page) {
    const fileName = `temp-file-${randomUUID()}.txt`;
    const tempDir = os.tmpdir();
    const filePath = path.join(tempDir, fileName);
    fs.writeFileSync(filePath, 'This is a test file content.');

    await page.goto('http://localhost:8080/');
    const fileChooserPromise = page.waitForEvent('filechooser');
    await page.click('#files-selector');
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(filePath);

    await page.$eval('#files-selector', (input) => {
        return (input as HTMLInputElement).files?.[0]?.name;
    });

    await page.click('input[type="submit"]');
    await page.waitForURL('**/view/**', { timeout: 5000 });

    return {
        fileName,
        filePath,
        cleanup: () => {
            if (fs.existsSync(filePath)) fs.unlinkSync(filePath);
        }
    };
}
