import * as path from 'node:path';
import { test } from '@playwright/test';

const ARTIFACT_DIR =
  '/Users/aobaiwaki/.gemini/antigravity-cli/brain/8ecc8ee2-c710-4933-8609-831f179266c5';

test('capture FeedbackModal comparing user answer vs correct colors', async ({
  page,
}) => {
  // Go to quiz page
  await page.goto('http://localhost:3000');

  // Wait for loading to finish and question to appear
  await page.waitForSelector('text=13th Single 制服', { timeout: 10000 });
  await page.waitForTimeout(1000);

  // 1. Click Left Penlight to open DonutRingModal
  await page.locator('div[title="タップして左手の色を選択"]').click();
  await page.waitForTimeout(500);

  // Select wrong color: レッド
  await page.locator('div[role="button"][aria-label="レッド"]').click();
  await page.waitForTimeout(500);

  // Select wrong color: グリーン -> triggers answer and opens FeedbackModal
  await page.locator('div[role="button"][aria-label="グリーン"]').click();
  await page.waitForTimeout(600);

  // 2. Capture FeedbackModal (Incorrect state with Your Answer vs Correct Colors comparison)
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-feedback-modal.png'),
    fullPage: true,
  });

  // 3. Click "次へ進む"
  await page.locator('button:has-text("次へ進む")').click();
  await page.waitForTimeout(800);

  // Capture next question screen
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-next-question.png'),
    fullPage: true,
  });
});
