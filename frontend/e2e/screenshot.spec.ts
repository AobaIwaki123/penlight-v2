import * as path from 'node:path';
import { test } from '@playwright/test';

const ARTIFACT_DIR =
  '/Users/aobaiwaki/.gemini/antigravity-cli/brain/8ecc8ee2-c710-4933-8609-831f179266c5';

test('capture transparent donut with visual penlights and bottom single-line feedback bar', async ({
  page,
}) => {
  // Go to quiz page
  await page.goto('http://localhost:3000');

  // Wait for loading to finish and question to appear
  await page.waitForSelector('text=13th Single 制服', { timeout: 10000 });
  await page.waitForTimeout(1000);

  // 1. Open Donut Modal by clicking left penlight on top of photo
  await page.locator('div[title="タップして左手の色を選択"]').click();
  await page.waitForTimeout(600);

  // Capture: Transparent Donut Ring (Oshi's face 100% visible behind!)
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-transparent-donut.png'),
    fullPage: true,
  });

  // 2. Click a color (e.g. イエロー) -> Left stick gets Yellow, focus shifts to Right stick
  const yellowColorBtn = page.locator(
    'div[role="button"][aria-label="イエロー"]',
  );
  await yellowColorBtn.click();
  await page.waitForTimeout(500);

  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-transparent-donut-step2.png'),
    fullPage: true,
  });

  // 3. Click wrong color for right hand (e.g. レッド) -> Modal closes, penlights shift to correct colors, single-line bottom bar appears!
  const redColorBtn = page.locator('div[role="button"][aria-label="レッド"]');
  await redColorBtn.click();
  await page.waitForTimeout(800);

  // Capture: Minimal single-line bottom bar + correct colors lit up on photo! Zero center block!
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-minimal-inline-feedback.png'),
    fullPage: true,
  });
});
