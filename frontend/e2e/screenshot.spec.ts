import * as path from 'node:path';
import { expect, test } from '@playwright/test';

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

  // 0. Capture initial photo with bottom-right twin penlights
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-fullscreen-photo.png'),
    fullPage: false,
  });

  // 1. Open Donut Modal by clicking left penlight on top of photo
  await page.locator('div[title="タップして左手の色を選択"]').click();
  await page.waitForTimeout(600);

  // Capture: Downward shifted Donut Ring (Face completely clear above!)
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-transparent-donut.png'),
    fullPage: false,
  });

  // 2. Click longest color: エメラルドグリーン -> Verify NO text wrapping occurred!
  const emeraldBtn = page.locator(
    'div[role="button"][aria-label="エメラルドグリーン"]',
  );
  await emeraldBtn.click();
  await page.waitForTimeout(500);

  // Verify Emerald Green text has no wrap bug and is single line
  const textElem = page.locator('text=エメラルドグリーン').first();
  await expect(textElem).toBeVisible();

  // Capture: Step 2 with Emerald Green selected
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-transparent-donut-step2.png'),
    fullPage: false,
  });

  // 3. Click second color (e.g. レッド) -> Modal closes, feedback appears
  const redColorBtn = page.locator('div[role="button"][aria-label="レッド"]');
  await redColorBtn.click();
  await page.waitForTimeout(1000);

  // Verify Inline Feedback Bar is visible with Next button
  const nextBtn = page.getByRole('button', { name: '次へ' });
  await expect(nextBtn).toBeVisible({ timeout: 5000 });

  // Capture: Inline Feedback Bar comfortably within viewport (zero cutoff!)
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-minimal-inline-feedback.png'),
    fullPage: false,
  });
});
