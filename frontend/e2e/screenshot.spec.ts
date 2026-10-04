import * as path from 'node:path';
import { test } from '@playwright/test';

const ARTIFACT_DIR =
  '/Users/aobaiwaki/.gemini/antigravity-cli/brain/8ecc8ee2-c710-4933-8609-831f179266c5';

test('capture mobile quiz layouts and interactions', async ({ page }) => {
  // Go to quiz page
  await page.goto('http://localhost:3000');

  // Wait for loading to finish and question to appear
  await page.waitForSelector('text=13th Single 制服', { timeout: 10000 });
  await page.waitForTimeout(1000);

  // 1. Capture Layout: Overlay (Default)
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-overlay.png'),
    fullPage: true,
  });

  // 2. Switch to Classic (旧版クラシック 縦並び)
  await page.locator('button[title="レイアウト変更"]').click();
  await page.waitForTimeout(300);
  await page.locator('text=旧版クラシック (縦並び)').click();
  await page.waitForTimeout(800);
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-classic.png'),
    fullPage: true,
  });

  // 3. Switch to Compact (コンパクト操作重視)
  await page.locator('button[title="レイアウト変更"]').click();
  await page.waitForTimeout(300);
  await page.locator('text=コンパクト操作重視').click();
  await page.waitForTimeout(800);
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-compact.png'),
    fullPage: true,
  });

  // 4. Switch back to Overlay and select a color
  await page.locator('button[title="レイアウト変更"]').click();
  await page.waitForTimeout(300);
  await page.locator('text=オーバーレイ没入型 ★').click();
  await page.waitForTimeout(500);

  // Click first color "イエロー"
  await page.locator('button:has-text("イエロー")').click();
  await page.waitForTimeout(400);
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-palette-selected.png'),
    fullPage: true,
  });

  // 5. Toggle Dark Mode
  await page.locator('button[title="テーマ切り替え"]').click();
  await page.waitForTimeout(400);
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-dark-mode.png'),
    fullPage: true,
  });
});
