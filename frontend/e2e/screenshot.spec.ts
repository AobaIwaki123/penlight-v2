import * as path from 'node:path';
import { test } from '@playwright/test';

const ARTIFACT_DIR =
  '/Users/aobaiwaki/.gemini/antigravity-cli/brain/8ecc8ee2-c710-4933-8609-831f179266c5';

test('capture donut ring palette and compare with grid', async ({ page }) => {
  // Go to quiz page
  await page.goto('http://localhost:3000');

  // Wait for loading to finish and question to appear
  await page.waitForSelector('text=13th Single 制服', { timeout: 10000 });
  await page.waitForTimeout(1000);

  // 1. Capture Donut Ring Input (Overlay Layout + Donut Ring)
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-donut-default.png'),
    fullPage: true,
  });

  // 2. Click a color on the Donut Ring (e.g., イエロー)
  const yellowColorBtn = page.locator(
    'div[role="button"][aria-label="イエロー"]',
  );
  await yellowColorBtn.click();
  await page.waitForTimeout(500);

  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-donut-selected.png'),
    fullPage: true,
  });

  // 3. Switch back to 15-Color Grid Input
  await page.locator('button[title="カラーパレット方式"]').click();
  await page.waitForTimeout(300);
  await page.locator('text=15色グリッド型').click();
  await page.waitForTimeout(500);

  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-grid.png'),
    fullPage: true,
  });
});
