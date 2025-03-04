import { defineConfig } from "playwright/test";

export default defineConfig({
    testDir: './tests',
    timeout: 30 * 1000,
    use: {
        headless: true
    },
    projects: [
        { name: 'chromium', use: { browserName: 'chromium' } },
        { name: 'firefox', use: { browserName: 'firefox' } },
        { name: 'webkit', use: { browserName: 'webkit' } },
    ],
});