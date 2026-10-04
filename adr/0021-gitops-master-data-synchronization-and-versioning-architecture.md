---
id: ADR-0021
title: GitOps マスタデータ同期アーキテクチャおよび複数画像・バージョン管理テーブルの採用
status: Accepted
scope: System
primary_category: DAT
categories: [DAT, ARC, OPS]
tags: [gitops, sqlite, typeid, normalization, pwa]
deciders: [user, ai]
date: 2026-10-04
---

# 0021. GitOps マスタデータ同期アーキテクチャおよび複数画像・バージョン管理テーブルの採用 (0021-gitops-master-data-synchronization-and-versioning-architecture.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-04

---

## 1. 背景と解決すべき課題 (Context & Problem)

本システムにおけるデータのライフサイクルと正本（SSoT）の所在について検討した結果、以下の課題が明確になった。

1. **アクセス頻度と更新特性の根本的差異**:
   - メンバー情報やペンライトカラー等のマスタデータは、**「ほぼ 100% 読み取り（Read）」** であり、更新は新期生加入や新曲発売など **「年に数回（低頻度）」** に限られる。
   - 一方、回答ログ（`ans_`）やユーザー情報（`usr_`）は、クイズ実行ごとに高頻度で書き込まれるトランザクションデータである。
2. **管理画面（CMS）自作による過剰設計 (YAGNI 違反)**:
   - 年に数回しか更新されないマスタのために、認証・認可付きの管理画面や画像アップロード API を自作するのは開発・保守コストが過大となる。
3. **本番と開発環境のデータ乖離（ドリフト）および上書き事故リスク**:
   - 本番 DB（R2）のみをマスタ正本とすると、ローカル環境との差分管理や Git での変更履歴の追跡（いつ誰が何を変えたか）がブラックボックス化する。
4. **同一メンバーにおける複数画像（衣装・グッズ等）の要件**:
   - シングル表題曲、期別衣装、ライブグッズ（個別タオル）など、同一メンバーに対して複数の画像が存在し、クイズ出題のバリエーションや PWA キャッシュの階層化のために 1対多 (1:N) で正規化管理する必要がある。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

### 1. マスタデータの GitOps 運用（Git を唯一の正本とする）
- **マスタデータ (SSoT)**: `seeds/data/*.json`（グループ、カラー、メンバー、画像）を正本とし、GitHub リポジトリで管理する。
- **更新フロー**: マスタの追加・修正はすべて GitHub PR を通して行い、CI（`./scripts/verify-all.sh`）で構文・型・テストを機械的に検証した上でマージする。
- **CD 自動反映（UPSERT）**: デプロイ時（またはサーバー起動時）、本番 SQLite に対して `INSERT INTO ... ON CONFLICT (id) DO UPDATE SET ...` を実行する。
  - 不変サロゲートキー（TypeID: `grp_`, `col_`, `mem_`, `img_`）により、回答ログとの外部キー結合を一切破壊することなく属性差分のみを無停止で安全に適用する。

### 2. メンバー画像の完全正規化 (`member_images` テーブル)
`members` テーブルから単一の `image_key` カラムを撤廃し、独立した `member_images` テーブルを新設する。

```sql
CREATE TABLE IF NOT EXISTS member_images (
    id TEXT PRIMARY KEY,                       -- img_<uuidv7> (TypeID)
    member_id TEXT NOT NULL,                    -- mem_<uuidv7> (FK)
    image_key TEXT NOT NULL UNIQUE,             -- img_<uuidv7>.webp (ADR-0008 不変配信)
    title TEXT NOT NULL,                        -- e.g. "13th Single アー写", "5thひな誕祭タオル"
    is_primary INTEGER NOT NULL DEFAULT 0,      -- 1: 代表画像, 0: バリエーション
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_member_images_member ON member_images(member_id, is_primary);
```

### 3. マスタバージョン管理テーブル (`master_versions`)
Local-First PWA（[ADR-0007](./0007-local-first-offline-pwa-architecture.md)）における `GET /api/v1/bootstrap` の差分検知（ETag）および CD 反映状態を追跡するため、バージョンテーブルを導入する。

```sql
CREATE TABLE IF NOT EXISTS master_versions (
    id TEXT PRIMARY KEY,                       -- 'current' 固定 (単一レコード)
    version TEXT NOT NULL,                     -- Git コミットハッシュまたはセマンティックバージョン
    updated_at TEXT NOT NULL
);
```

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

| 評価項目 | クラウドCMS型 (本番直編集) | GitOps 型 [採択] |
|---|---|---|
| **管理画面の開発工数** | 甚大 (UI・認証・API必須) | **ゼロ (JSON編集+PRのみ)** |
| **変更履歴の監査性** | 低 (バイナリWALのみ) | **最高 (Gitコミット・PRログ)** |
| **環境間のデータ一致** | 乖離しやすい (要Dump同期) | **完全一致 (コードと同等管理)** |
| **誤操作のロールバック** | 難 (PITR手動オペ) | **即時 (git revert)** |
| **複数画像の柔軟性** | 実装に依存 | **完全正規化 (1:N)** |

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響 (Positive)
- **CMS 開発の完全排除**: 複雑な管理画面を作ることなく、安全なマスタ更新サイクルを確立。
- **事故の撲滅**: デプロイ時に古いシード値で本番が巻き戻るリスクを排除し、UPSERT による安全なインプレース更新を実現。
- **PWA との親和性**: `master_versions` により、オフライン PWA クライアントがミリ秒でマスタ更新を検知可能。

### 留意点と対策 (Negative & Mitigation)
- **JOIN クエリの発生**:
  - 画像取得に `member_images` との JOIN が必要となるが、`idx_member_images_member` インデックスにより SQLite WAL 上で 0.1ms 未満で実行可能。
