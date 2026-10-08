import { expect, test } from '@playwright/test';
import type { SubmitMetadataEditProposalRequest } from '../src/types/generated';

const groupId = 'grp_019245a1000070008000000000000001';
const colorIds = [1, 2, 3].map(
  (number) => `col_019245a100007000800000000000000${number}`,
);
const photoTypeIds = [1, 2].map(
  (number) => `pht_019245a100007000800000000000000${number}`,
);
const members = ['テスト 一花', 'テスト 二葉'].map((name, index) => ({
  id: `mem_019245a100007000800000000000000${index + 1}`,
  group_id: groupId,
  family_name: name.split(' ')[0],
  given_name: name.split(' ')[1],
  generation: index + 2,
  status: 'active',
  metadata_revision: 4,
  penlight: {
    left_color_id: colorIds[0],
    right_color_id: colorIds[1],
    ordered: true,
  },
  images: [1, 2].map((number) => ({
    id: `img_019245a10000700080000000000000${index}${number}`,
    image_key: `img_019245a10000700080000000000000${index}${number}.webp`,
    photo_type_id: photoTypeIds[number - 1],
    photo_type: { name: `衣装${number}` },
    is_primary: number === 1,
  })),
}));

test.beforeEach(async ({ page }) => {
  await page.route('**/api/v1/sync/bootstrap?include_graduated=true', (route) =>
    route.fulfill({
      json: {
        series: [{ id: 'ser_1', name: '坂道' }],
        groups: [{ id: groupId, series_id: 'ser_1', name: 'テストグループ' }],
        members,
        colors: ['ブルー', 'イエロー', 'グリーン'].map((name, index) => ({
          id: colorIds[index],
          name,
          hex_code: ['#228be6', '#fcc419', '#40c057'][index],
          display_order: index,
        })),
        photo_types: photoTypeIds.map((id, index) => ({
          id,
          group_id: groupId,
          name: `衣装${index + 1}`,
        })),
      },
    }),
  );
  await page.route('**/images/*', (route) =>
    route.fulfill({
      contentType: 'image/svg+xml',
      body: '<svg xmlns="http://www.w3.org/2000/svg" width="400" height="600"><rect width="100%" height="100%" fill="#495057"/></svg>',
    }),
  );
  await page.addInitScript(() => {
    setInterval(() => {
      document.querySelectorAll('nextjs-portal').forEach((el) => {
        el.remove();
      });
    }, 50);
  });
});

test('edit answers in place, navigate both ways, and retry the same submission', async ({
  page,
}) => {
  const requests: SubmitMetadataEditProposalRequest[] = [];
  await page.route('**/api/v1/members/*/metadata-edit-proposals', (route) => {
    const request = route.request().postDataJSON();
    requests.push(request);
    return route.fulfill(
      requests.length === 1
        ? { status: 503, json: { detail: 'テスト再送' } }
        : { json: { ...request, member_id: members[0].id, status: 'pending' } },
    );
  });

  await page.goto('/edit');
  await expect(page.getByText('テスト 一花')).toBeVisible();
  await expect(
    page.getByRole('button', { name: '公式回答と一致' }),
  ).toBeDisabled();

  // 1. 期生と在籍ステータスを変更
  await page.getByText('2期生').click();
  const infoModal = page.getByRole('dialog', {
    name: 'テスト 一花 の情報を編集',
  });
  await expect(infoModal).toBeVisible();
  await infoModal.getByRole('textbox', { name: '期生' }).fill('4');
  await infoModal.getByText('卒業', { exact: true }).click();
  await infoModal.getByRole('button', { name: '完了' }).click();

  // 2. ペンライトカラーを変更
  await page.locator('[title="タップして左手の色を選択"]').click();
  const donut = page.getByRole('dialog');
  await expect(donut).toBeVisible();
  await donut.getByRole('button', { name: 'グリーン' }).click();
  await expect(donut).not.toBeVisible();

  // 3. 次のメンバーへ移動して戻る
  await page.getByRole('button', { name: '次へ' }).click();
  await expect(page.getByText('テスト 二葉')).toBeVisible();
  await page.getByRole('button', { name: '前へ' }).click();
  await expect(page.getByText('テスト 一花')).toBeVisible();
  await expect(page.getByText('4期生')).toBeVisible();

  // 4. 差分確認と提案送信 (1回目: 503エラー)
  await page.getByRole('button', { name: /項目の修正を提案/ }).click();
  const diffModal = page.getByRole('dialog', { name: '提案内容の確認' });
  await expect(diffModal).toBeVisible();
  await diffModal.getByRole('button', { name: 'この内容で提案を送信' }).click();
  await expect(diffModal.getByText('テスト再送')).toBeVisible();

  // 次へ行って戻ってもドラフトと再送が維持されること
  await diffModal.getByRole('button', { name: '閉じる' }).click();
  await page.getByRole('button', { name: '次へ' }).click();
  await expect(page.getByText('テスト 二葉')).toBeVisible();
  await page.getByRole('button', { name: '前へ' }).click();
  await expect(page.getByText('テスト 一花')).toBeVisible();

  // 5. 再送 (2回目: 成功)
  await page.getByRole('button', { name: /項目の修正を提案/ }).click();
  const retryModal = page.getByRole('dialog', { name: '提案内容の確認' });
  await retryModal
    .getByRole('button', { name: 'この内容で提案を送信' })
    .click();
  await expect(
    page.getByRole('button', { name: '提案送信済み (承認待ち)' }),
  ).toBeVisible();

  expect(requests).toHaveLength(2);
  expect(requests[1]).toEqual(requests[0]);
  expect(requests[0]).toMatchObject({
    base_revision: 4,
    changes: {
      generation: { before: 2, after: 4 },
      status: { before: 'active', after: 'graduated' },
      penlight: {
        before: members[0].penlight,
        after: { ...members[0].penlight, left_color_id: colorIds[2] },
      },
    },
  });
});

test('opens directly to target member when member_id query param is provided', async ({
  page,
}) => {
  const targetMember = members[1]; // テスト 二葉
  // beforeEach already sets up the bootstrap mock for include_graduated=true
  await page.goto(`/edit?member_id=${targetMember.id}`);

  // 二葉が最初に表示されていること
  await expect(page.getByText('テスト 二葉')).toBeVisible();
  await expect(page.getByText('3期生')).toBeVisible();
});

test('supports swipe navigation with draft preservation and list boundary resistance', async ({
  page,
}) => {
  await page.goto('/edit');
  await expect(page.getByText('テスト 一花')).toBeVisible();

  // 1. 一花で期生を変更
  await page.getByText('2期生').click();
  const infoModal = page.getByRole('dialog', {
    name: 'テスト 一花 の情報を編集',
  });
  await expect(infoModal).toBeVisible();
  await infoModal.getByRole('textbox', { name: '期生' }).fill('5');
  await infoModal.getByRole('button', { name: '完了' }).click();
  await expect(page.getByText('5期生')).toBeVisible();

  // 2. 左スワイプで次へ移動 (一花 -> 二葉)
  const card = page.locator('text=テスト 一花').first();
  const box = await card.boundingBox();
  if (box) {
    const startX = box.x + box.width / 2;
    const startY = box.y + box.height / 2;
    await page.mouse.move(startX, startY);
    await page.mouse.down();
    await page.mouse.move(startX - 180, startY, { steps: 5 });
    await page.mouse.up();
  }

  // 二葉へ切り替わっていること
  await expect(page.getByText('テスト 二葉')).toBeVisible();

  // 3. 末尾でさらに左スワイプしても循環しないこと (二葉にとどまる)
  const card2 = page.locator('text=テスト 二葉').first();
  const box2 = await card2.boundingBox();
  if (box2) {
    const startX = box2.x + box2.width / 2;
    const startY = box2.y + box2.height / 2;
    await page.mouse.move(startX, startY);
    await page.mouse.down();
    await page.mouse.move(startX - 180, startY, { steps: 5 });
    await page.mouse.up();
  }
  await expect(page.getByText('テスト 二葉')).toBeVisible();

  // 4. 右スワイプで前へ戻る (二葉 -> 一花)
  if (box2) {
    const startX = box2.x + box2.width / 2;
    const startY = box2.y + box2.height / 2;
    await page.mouse.move(startX, startY);
    await page.mouse.down();
    await page.mouse.move(startX + 180, startY, { steps: 5 });
    await page.mouse.up();
  }

  // 一花へ戻り、変更した下書き（5期生）が保持されていること
  await expect(page.getByText('テスト 一花')).toBeVisible();
  await expect(page.getByText('5期生')).toBeVisible();
});
