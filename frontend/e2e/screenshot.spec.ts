import * as path from 'node:path';
import { expect, test } from '@playwright/test';

const ARTIFACT_DIR =
  '/Users/aobaiwaki/.gemini/antigravity-cli/brain/8ecc8ee2-c710-4933-8609-831f179266c5';

test('capture anchor picker, balanced donut modal, and luminous feedback bar', async ({
  page,
}) => {
  // Go to quiz page
  await page.goto('http://localhost:3000');

  // Wait for loading to finish and question to appear
  await page.waitForSelector('text=13th Single 制服', { timeout: 10000 });
  await page.waitForTimeout(800);

  // 1. [C案: 手元アンカー追従型] 写真右下の左ペンライトをタップ
  await page.locator('div[title="タップして左手の色を選択"]').click();
  await page.waitForTimeout(500);

  // Capture: Anchor Color Picker floating right above the penlight! Face 100% unobstructed!
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-anchor-picker.png'),
    fullPage: false,
  });

  // 2. Select longest color: エメラルドグリーン
  const emeraldBtn = page.locator(
    'div[role="button"][aria-label="エメラルドグリーン"]',
  );
  await emeraldBtn.click();
  await page.waitForTimeout(500);

  // 3. Select second color: レッド -> triggers answer submission
  const redBtn = page.locator('div[role="button"][aria-label="レッド"]');
  await redBtn.click();
  await page.waitForTimeout(800);

  // Verify Luminous Feedback Bar appears
  const nextBtn = page.getByRole('button', { name: '次へ' });
  await expect(nextBtn).toBeVisible({ timeout: 5000 });

  // Capture: Luminous Feedback Bar with glowing twin pills!
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-luminous-feedback.png'),
    fullPage: false,
  });

  // Advance to next question
  await nextBtn.click();
  await page.waitForTimeout(800);

  // 4. Switch input mode to "洗練ドーナツモーダル" via header menu
  const paletteMenuBtn = page.locator('button[title="カラーパレット方式"]');
  await paletteMenuBtn.click();
  await page.waitForTimeout(400);

  const donutMenuItem = page.locator('text=洗練ドーナツモーダル');
  await donutMenuItem.click();
  await page.waitForTimeout(500);

  // 5. Open balanced donut modal
  await page.locator('div[title="タップして左手の色を選択"]').click();
  await page.waitForTimeout(600);

  // Capture: Balanced Donut Modal (slimmer, face clear, background sticks faded)
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-balanced-donut.png'),
    fullPage: false,
  });
});
