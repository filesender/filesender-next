import { test, expect } from '@playwright/test';
import { uploadTestFile } from './fixtures/upload-helper';

test('should upload a file', async ({ page }) => {
    const { cleanup } = await uploadTestFile(page);
    try {
        const pageBody = await page.content();
        expect(pageBody).toContain("2048 bytes");
    } finally {
        cleanup();
    }
});
