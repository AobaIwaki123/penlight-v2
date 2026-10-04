---
id: ADR-0023
title: マスタデータ差分同期保証および整合性機械検証スクリプトの採用
status: Accepted
scope: System
primary_category: DAT
categories: [DAT, OPS, DEV]
tags: [sqlite, gitops, seed, verification, typeid]
deciders: [user, ai]
date: 2026-10-04
---

# 0023. マスタデータ差分同期保証および整合性機械検証スクリプトの採用 (0023-master-data-sync-guarantee-and-verification.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-04

---

## 1. 背景と解決すべき課題 (Context & Problem)

[ADR-0021](./0021-gitops-master-data-synchronization-and-versioning-architecture.md) で策定した GitOps マスタ同期の実装・運用において、以下の 3 つの課題が発生した。

1. **`INSERT OR IGNORE` による旧画像の幽霊レコード残留と 404 エラー**:
   - `member_images` の主キー（`id`）および画像キー（`image_key`）は画像 URL の SHA-1 ハッシュから導出される。
   - 画像 URL の更新時に新しい TypeID が生成されるが、従来のシード投入文が `INSERT OR IGNORE` だったため、旧 URL のレコードが DB 内に残留。
   - 単一メンバーに対して `is_primary = 1` が複数存在する状態となり、メンバー一覧取得時の `LEFT JOIN member_images` で旧画像キーが優先参照され、画像配信エンドポイント（`/images/{key}`）で 404 が発生した。

2. **起動時シード投入条件の不備によるマスタ未反映**:
   - サーバー起動時のシード適用判定が `SELECT COUNT(*) FROM groups == 0`（完全空 DB のみ）となっていた。
   - その結果、Git 上でシード（新メンバー、カラー名、画像 URL）を更新してサーバーを再起動しても、稼働中 DB に差分が一切反映されなかった。

3. **対話型 ad-hoc 調査による LLM トークンの過剰消費**:
   - DB 内部の重複行や画像キーの一致状況、API の ETag 健全性を確認するために、都度エージェントが対話型 SQL クエリや使い捨ての調査スクリプトを実行し、大量のプロンプト／出力トークンとターン数を浪費していた。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

### 2.1 完全 UPSERT ＆ 画像リセット規約 (`scripts/build_seed.go`)

マスタデータのシード生成において、属性更新と画像整合性を担保する以下の SQL 生成規約を採用する。

1. **基本マスタの UPSERT 化**:
   - `groups`, `colors`, `photo_types`, `members` はすべて `INSERT INTO ... ON CONFLICT (id) DO UPDATE SET ...` を出力する。
   - TypeID による決定論的 ID 生成（[ADR-0001](./0001-surrogate-key-typeid-uuidv7.md)）により、回答ログ（`answer_logs`）の外部キー制約を破壊することなく、属性差分（名称、カラーコード、期別等）を安全に適用する。
2. **`member_images` の事前クリアと再投入**:
   - 画像 URL の変更に伴い主キー（`img_<uuid>`）が変化するため、`member_images` 投入ブロックの直前に `DELETE FROM member_images;` を実行する。
   - これにより、旧 URL の幽霊レコードを完全に排除し、各メンバーの代表画像（`is_primary = 1`）の一意性を保証する。

### 2.2 バージョン駆動型起動同期 (`master_versions`)

静的な空 DB 判定を撤廃し、バージョン比較による起動時自動同期を実装する。

```go
// pkg/model/master_version.go
const CurrentMasterVersion = "2026.10.04-4"

// pkg/server/server.go
var currentVersion string
_ = repo.DB().QueryRow("SELECT version FROM master_versions WHERE id = 'current';").Scan(&currentVersion)
if currentVersion != model.CurrentMasterVersion {
    // 最新の seeds/seed.sql をトランザクション実行して同期
}
```

- クライアント PWA は `GET /api/v1/sync/bootstrap` の `ETag: "2026.10.04-4"` により、起動時のマスタ更新を即座に検知する（[ADR-0007](./0007-local-first-offline-pwa-architecture.md)）。

### 2.3 機械的整合性検証スクリプト (`scripts/verify_master.go`) の標準配備

AI エージェントおよび開発者が対話型クエリでトークンを浪費することを防ぐため、単一コマンド（約 1 秒）で DB・サーバー健全性を検証する専用スクリプトを配備する。

```bash
go run ./scripts/verify_master.go
```

**自動検証チェックリスト**:
1. `master_versions.version` と Go 側定数（`model.CurrentMasterVersion`）の一致。
2. `member_images` の代表画像（`is_primary = 1`）の重複ゼロ（`COUNT(*) == COUNT(DISTINCT member_id)`）。
3. マスタテーブル件数（グループ、カラー、メンバー、アクティブメンバー数）の妥当性。
4. SQLite 内の全 `image_key` が `image_sources.json` に実在することの検証。
5. （サーバー稼働時）`GET /api/v1/sync/bootstrap` の ETag 応答検証。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### マスタデータ同期・検証方式の比較

| 評価項目 | 毎回全DB再生成 | 手動 Migration / パッチ | バージョン駆動 UPSERT ＋ 検証スクリプト [採択] |
|---|---|---|---|
| **回答ログ（トランザクション）の保全** | 不可（DB再作成で全消失） | 可能だが人的ミス多発 | **完全保全（TypeID による外部キー保護）** |
| **ゴースト行・重複画像の防止** | 可能 | 漏れやすい | **完全排除（`DELETE FROM member_images`）** |
| **起動・適用オーバーヘッド** | 甚大（全テーブル再作成） | 人的オペレーション必須 | **極小（差分検知時は数ミリ秒で適用完了）** |
| **AI エージェントのトークン消費** | 中 | 甚大（調査の往復） | **最小（スクリプト 1 実行・約 1 秒で判定）** |

---

## 4. 得られる効果と運用規約 (Consequences & Enforcement)

### 4.1 ポジティブな影響 (Positive)
- **404 エラー・表示不整合の根絶**: 代表画像の重複登録を根本から防止し、画像キーの不整合によるクライアント表示エラーを排除。
- **GitOps の自動適用**: シード JSON の更新をマージ後、サーバーが再起動するだけで本番 SQLite に自動同期（手動マイグレーション不要）。
- **AI 開発におけるトークン消費削減**: エージェントが手動で複数の SQL やワンライナーを実行して状態を突き合わせる必要がなくなり、単一スクリプトの exit code とサマリで即座に健全性を確認可能。

### 4.2 運用規約 (Enforcement)
- **マスタ更新時の検証必須化**: `seeds/data/*.json` を変更した際は、`scripts/build_seed.go` でシードを再生成後、必ず `go run ./scripts/verify_master.go` が全項目パスすることを確認する。
- **手動クエリによるトークン浪費の禁止**: AI エージェントは DB 内部状態の調査にあたり、ad-hoc な SQL クエリや使い捨てスクリプトを連打してはならず、本スクリプトの実行結果を第一の判断材料とする。
