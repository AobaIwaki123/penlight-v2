import * as path from 'node:path';
import { expect, test } from '@playwright/test';

const ARTIFACT_DIR =
  '/Users/aobaiwaki/.gemini/antigravity-cli/brain/8ecc8ee2-c710-4933-8609-831f179266c5';

test('capture persistent bottom-right color wheel and glowing corner penlights', async ({
  page,
}) => {
  // Go to quiz page
  await page.goto('http://localhost:3000');

  // Wait for loading to finish and question to appear
  await page.waitForSelector('text=13th Single 制服', { timeout: 10000 });
  await page.waitForTimeout(800);

  // 1. [常時表示カラーピッカー & 左上ペンライト発光オーラ]
  // 画面を開いた瞬間から右下にカラーピッカーが常時表示！左上ペンライトが青く発光オーラ！
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-anchor-picker.png'),
    fullPage: false,
  });

  // 2. Select longest color for left hand: エメラルドグリーン
  const emeraldBtn = page.locator(
    'div[role="button"][aria-label="エメラルドグリーン"]',
  );
  await emeraldBtn.click();
  await page.waitForTimeout(600);

  // 3. Select second color for right hand: レッド -> triggers answer submission after 280ms
  const redBtn = page.locator('div[role="button"][aria-label="レッド"]');
  await redBtn.click();
  await page.waitForTimeout(1000);

  // Verify Luminous Feedback Bar appears
  const nextBtn = page.getByRole('button', { name: '次へ' });
  await expect(nextBtn).toBeVisible({ timeout: 5000 });

  // Capture: Luminous Feedback Bar with glowing twin pills & lit-up top corners!
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-luminous-feedback.png'),
    fullPage: false,
  });
});
