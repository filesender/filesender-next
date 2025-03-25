import { test, expect } from '@playwright/test';
import fs from 'fs';
import { randomUUID } from 'crypto';

test('should return a non successful response', async ({ page }) => {
    const response = await page.goto('http://localhost:8080/upload/this doesnt exist');

    expect(response?.status()).toBe(400);
    expect(await response?.text()).toContain("Transfer ID is invalid");
});

test('should upload a file', async ({ page }) => {
    // Generate temp file
    const filePath = `temp-file-${randomUUID()}.txt`;
    fs.writeFileSync(filePath, 'This is a test file content.'); // 28 bytes

    await page.goto('http://localhost:8080/');

    const fileChooserPromise = page.waitForEvent('filechooser');
    await page.click('#files-selector');
    const fileChooser = await fileChooserPromise;

    await fileChooser.setFiles(filePath);

    const uploadedFileName = await page.$eval('#files-selector', (input) => {
        return (input as HTMLInputElement).files?.[0]?.name;
    });
    expect(uploadedFileName).toBe(filePath);

    await page.click('input[type="submit"]');
    await page.waitForURL('**\/upload\/**',{
        timeout: 5000
    });

    const pageBody = await page.content();
    expect(pageBody).toContain("28 bytes");
    expect(pageBody).toContain("1 file");

    fs.unlinkSync(filePath);
});
