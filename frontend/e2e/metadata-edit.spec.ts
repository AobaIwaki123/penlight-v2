import { expect, type Page, test } from '@playwright/test';
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
        groups: [{ id: groupId, name: 'テストグループ' }],
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
});

async function swipe(page: Page, direction: 'left' | 'right') {
  const surface = await page.getByTestId('member-swipe-surface').boundingBox();
  if (!surface) throw new Error('Swipe surface is not visible');
  const left = surface.x + 36;
  const right = surface.x + surface.width - 36;
  const start = direction === 'left' ? right : left;
  const end = direction === 'left' ? left : right;
  const y = surface.y + 120;
  const session = await page.context().newCDPSession(page);
  await session.send('Input.dispatchTouchEvent', {
    type: 'touchStart',
    touchPoints: [{ x: start, y }],
  });
  for (let step = 1; step <= 6; step++) {
    await session.send('Input.dispatchTouchEvent', {
      type: 'touchMove',
      touchPoints: [{ x: start + ((end - start) * step) / 6, y }],
    });
  }
  await session.send('Input.dispatchTouchEvent', {
    type: 'touchEnd',
    touchPoints: [],
  });
  await session.detach();
}

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
  await expect(page.getByTestId('current-member-name')).toHaveText(
    'テスト 一花',
  );
  await expect(
    page.getByRole('button', { name: 'この回答を提案' }),
  ).toBeDisabled();
  await swipe(page, 'right');
  await expect(page.getByTestId('current-member-name')).toHaveText(
    'テスト 一花',
  );
  await page.getByRole('textbox', { name: '期生', exact: true }).fill('4');
  await page.getByRole('textbox', { name: '在籍状態' }).click();
  await page.getByRole('option', { name: '休業中', exact: true }).click();

  await page.locator('[title="タップして左手の色を選択"]').click();
  const donut = page.getByRole('dialog');
  await expect(donut).toBeVisible();
  await donut.getByRole('button', { name: 'グリーン', exact: true }).click();
  await expect(donut).not.toBeVisible();
  expect(requests).toHaveLength(0);

  await swipe(page, 'left');
  await expect(page.getByTestId('current-member-name')).toHaveText(
    'テスト 二葉',
  );
  await swipe(page, 'left');
  await expect(page.getByTestId('current-member-name')).toHaveText(
    'テスト 二葉',
  );
  await swipe(page, 'right');
  await expect(page.getByTestId('current-member-name')).toHaveText(
    'テスト 一花',
  );
  await expect(
    page.getByRole('textbox', { name: '期生', exact: true }),
  ).toHaveValue('4期生');
  await expect(
    page
      .getByTestId('member-swipe-surface')
      .getByText('グリーン', { exact: true }),
  ).toBeVisible();

  await page.getByRole('button', { name: '写真を選ぶ' }).click();
  await page.getByRole('button', { name: '代表写真: 衣装2' }).click();
  await expect(
    page.getByRole('button', { name: '代表写真: 衣装2' }),
  ).not.toBeVisible();
  await expect(
    page.getByRole('textbox', { name: '写真の衣装タグ' }),
  ).toHaveValue('衣装2');
  await page.getByRole('textbox', { name: '写真の衣装タグ' }).click();
  await page.getByRole('option', { name: '衣装1', exact: true }).click();

  await page.getByRole('button', { name: 'この回答を提案' }).click();
  await expect(page.getByText('テスト再送')).toBeVisible();
  await swipe(page, 'left');
  await expect(page.getByTestId('current-member-name')).toHaveText(
    'テスト 二葉',
  );
  await swipe(page, 'right');
  await expect(page.getByTestId('current-member-name')).toHaveText(
    'テスト 一花',
  );
  await page.getByRole('button', { name: '同じ回答を再送' }).click();
  await expect(
    page.getByRole('button', { name: '送信済み・承認待ち' }),
  ).toBeDisabled();
  expect(requests).toHaveLength(2);
  expect(requests[1]).toEqual(requests[0]);
  expect(requests[0]).toMatchObject({
    base_revision: 4,
    changes: {
      generation: { before: 2, after: 4 },
      status: { before: 'active', after: 'hiatus' },
      penlight: {
        before: members[0].penlight,
        after: { ...members[0].penlight, left_color_id: colorIds[2] },
      },
      primary_image_id: {
        before: members[0].images[0].id,
        after: members[0].images[1].id,
      },
      image_photo_types: [
        {
          image_id: members[0].images[1].id,
          before: photoTypeIds[1],
          after: photoTypeIds[0],
        },
      ],
    },
  });
});

test('show a local fallback when the photo upstream fails', async ({
  page,
}) => {
  await page.route('**/images/*', (route) =>
    route.fulfill({ status: 502, body: 'upstream image fetch error' }),
  );
  await page.goto('/edit');
  const photo = page.getByRole('img', { name: 'テスト 一花' });
  await expect(photo).toHaveAttribute('src', /^data:image\/svg\+xml,/);
  await expect
    .poll(() => photo.evaluate((image: HTMLImageElement) => image.naturalWidth))
    .toBeGreaterThan(0);
  await expect(
    page.getByRole('button', { name: 'この回答を提案' }),
  ).toBeDisabled();
});
