import * as path from 'node:path';
import { test } from '@playwright/test';

const ARTIFACT_DIR =
  '/Users/aobaiwaki/.gemini/antigravity-cli/brain/8ecc8ee2-c710-4933-8609-831f179266c5';

test('capture full-screen photo layout and donut ring modal on penlight tap', async ({
  page,
}) => {
  // Go to quiz page
  await page.goto('http://localhost:3000');

  // Wait for loading to finish and question to appear
  await page.waitForSelector('text=13th Single 制服', { timeout: 10000 });
  await page.waitForTimeout(1000);

  // 1. Capture Full-Screen Photo Mode (No palette at bottom! 100% photo immersion)
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-fullscreen-photo.png'),
    fullPage: true,
  });

  // 2. Click the Left Penlight on top of the photo to open Donut Ring Modal
  await page.locator('div[title="タップして左手の色を選択"]').click();
  await page.waitForTimeout(600);

  // Capture Modal Overlay with Donut Ring
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-modal-donut-open.png'),
    fullPage: true,
  });

  // 3. Select a color on the Donut Ring (e.g. イエロー)
  const yellowColorBtn = page.locator(
    'div[role="button"][aria-label="イエロー"]',
  );
  await yellowColorBtn.click();
  await page.waitForTimeout(600);

  // Capture step 2 (Left hand set to Yellow, now selecting Right hand)
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-modal-donut-step2.png'),
    fullPage: true,
  });

  // 4. Select right hand color (e.g. レッド) -> modal closes and returns to full photo
  const redColorBtn = page.locator('div[role="button"][aria-label="レッド"]');
  await redColorBtn.click();
  await page.waitForTimeout(800);

  // Capture final answered state on full-screen photo
  await page.screenshot({
    path: path.join(ARTIFACT_DIR, 'screenshot-fullscreen-answered.png'),
    fullPage: true,
  });
});
