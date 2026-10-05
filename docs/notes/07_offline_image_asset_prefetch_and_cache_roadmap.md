# 07. 画像アセットの完全オフライン・バックグラウンドプリフェッチ & CacheStorage 実装ロードマップ

- **ステータス**: 検討中 (Draft Roadmap)
- **日付**: 2026-10-05
- **関連 ADR**:
  - [ADR-0007: ライブ会場での完全動作を保証する Local-First / オフライン PWA アーキテクチャの採用](../../adr/0007-local-first-offline-pwa-architecture.md)
  - [ADR-0008: 画像アセットの永久不変キャッシュ (RFC 8246 immutable) とゼロパージ運用の採用](../../adr/0008-immutable-image-caching-and-zero-purge.md)

---

## 1. 背景と課題（現状の構造的欠落）

ADR-0007 では、電波が極度に混雑するライブ会場での利用を想定し、「端末内データによる完全ローカル出題」を規定している。
また ADR-0008 により、画像レスポンスには `Cache-Control: public, max-age=31536000, immutable` が付与されている。

しかし、現在のフロントエンド実装には以下の構造的課題が存在する：

1. **未表示画像のキャッシュ欠落**:
   - `/api/v1/sync/bootstrap` で取得しているのは画像キー（URL）を含むメタデータ（JSON）のみ。
   - ブラウザ標準の HTTP キャッシュは「画面に表示された画像（GET リクエストが走ったもの）」しか保持しない。
2. **圏外環境での画像脱落**:
   - 初回起動時やオンライン時に表示されなかったメンバー・別衣装の画像は端末内に存在しないため、ライブ会場の圏外で初めて出題された瞬間に画像ロードがタイムアウトし、表示が崩れる。
   - 衣装ランダム出題の導入により、この問題の発生確率が大幅に増大する。

---

## 2. 要件仕様

1. **バックグラウンド非同期一括プリフェッチ**:
   - アプリ起動時のマスタデータ取得（`fetchBootstrapData`）直後、全メンバー・全衣装の WebP 画像を非同期にダウンロードしてローカルキャッシュに蓄積する。
   - UI スレッドおよびクイズ操作を一切ブロックしないこと。
2. **Cache API (CacheStorage) による明示的永続化**:
   - ブラウザの揮発性メモリキャッシュに頼らず、`CacheStorage`（例: `penlight-images-v1`）に永続保存する。
3. **データ転送量の最小化**:
   - 1枚あたり 15〜25KB（WebP）のため、全メンバー（約70〜100名）× 複数衣装でも総データ量は 2〜4MB 程度に収まる。
   - すでにキャッシュ済みの画像は二重ダウンロードしない。

---

## 3. 技術設計

### 3.1. プリフェッチ処理の実装 (`frontend/src/features/quiz/api/prefetch.ts`)

```ts
export async function prefetchImages(members: Member[]): Promise<void> {
  if (typeof window === 'undefined' || !('caches' in window)) return;

  const imageUrls = Array.from(
    new Set(
      members.flatMap((m) =>
        (m.images || []).map((img) => getImageUrl(img.image_key))
      ).filter(Boolean)
    )
  );

  const cache = await caches.open('penlight-images-v1');

  // 未キャッシュの画像のみを抽出して並行ダウンロード
  await Promise.allSettled(
    imageUrls.map(async (url) => {
      const match = await cache.match(url);
      if (!match) {
        await cache.add(url);
      }
    })
  );
}
```

### 3.2. Service Worker による Cache-First インターセプト (`public/sw.js`)
Service Worker を導入し、`/images/*` へのリクエストを横取りして `CacheStorage` から 0ms で返却する。

```js
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);
  if (url.pathname.startsWith('/images/')) {
    event.respondWith(
      caches.open('penlight-images-v1').then(async (cache) => {
        const cached = await cache.match(event.request);
        if (cached) return cached;
        const response = await fetch(event.request);
        if (response.ok) {
          cache.put(event.request, response.clone());
        }
        return response;
      })
    );
  }
});
```

---

## 4. 実装ステップと検証計画

1. **ステップ 1: バックグラウンドプリフェッチ関数の実装**
   - `prefetch.ts` を実装し、`QuizContainer` の初期化完了後に非同期トリガー。
2. **ステップ 2: Service Worker の登録と配信ハンドラーの配置**
   - Next.js 静的エクスポート（`output: 'export'`）と共存する `public/sw.js` の配置と登録。
3. **ステップ 3: 検証**
   - Chrome DevTools の「Network: Offline」および「Application > Cache storage」タブで、未出題の全衣装画像が保存され、オフラインでも即時表示されることを検証。
   - `make verify-ai` の静的解析通過確認。
