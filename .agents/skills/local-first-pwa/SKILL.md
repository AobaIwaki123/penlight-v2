---
name: local-first-pwa
description: ライブ会場での圏外完結出題（Local-First）、電波復旧時バッチ同期、Next.js静的エクスポート（単一バイナリ）、不変画像配信を統制する。
---

# Local-First / 単一バイナリ規約 (local-first-pwa)

> **管轄 ADR**: [ADR-0002](../../../adr/0002-backend-go-architecture.md), [ADR-0007](../../../adr/0007-local-first-offline-pwa-architecture.md), [ADR-0008](../../../adr/0008-immutable-image-caching-and-zero-purge.md), [ADR-0011](../../../adr/0011-directory-structure-and-responsibility-boundaries.md), [ADR-0018](../../../adr/0018-quiz-candidate-pool-filtering-architecture.md), [ADR-0019](../../../adr/0019-quiz-format-strategy-and-color-palette-architecture.md), [ADR-0020](../../../adr/0020-quiz-target-member-selection-strategy.md)

本スキルは、ライブ会場での圏外動作を保証するクライアント設計、単一バイナリ配信、および画像不変キャッシュを統制する。

## 絶対遵守ルール (Invariants)

1. **クイズ設問生成・採点の端末内完結 (Local-First)**:
   - マスタデータ（グループ・メンバー・色情報）は初回に IndexedDB へ保存する。
   - クイズプレイ時の設問生成（出題メンバー選出 ADR-0020、解答形式 ADR-0019、誤答選定 ADR-0009）および正誤判定はブラウザ内の純粋関数で完結させ、毎問サーバー問い合わせを行わないこと。
2. **回答ログのバッチ同期と冪等性**:
   - 圏外時の回答ログは IndexedDB に蓄積し、回線復旧時に `POST /api/v1/quiz/answers/batch` で一括送信する。
   - サーバー側は `INSERT OR IGNORE` で処理し、多重送信によるデータ不整合を防止する。
3. **フロントエンドの完全静的エクスポート**:
   - Next.js は `output: 'export'` でビルドし、生成された静的ファイルを Go バイナリ（`embed.FS`）に内包して単一 Pod で配信する。
   - Next.js 側に API Routes や Server Actions 等の Node.js 依存バックエンド処理を作成してはならない。
4. **画像アセットの上書き禁止 (Zero-Purge)**:
   - 画像写真は `mem_<uuid>.webp` として命名し、`Cache-Control: public, max-age=31536000, immutable` で配信する。
   - 写真差し替え時は同じファイル名で上書きせず、新しい UUID で新規ファイルを作成すること。
