---
id: ADR-0035
title: メタデータ確認・編集 API および GitOps 逆同期アーキテクチャの採用
status: Accepted
scope: System
primary_category: DAT
categories: [DAT, APP, DEV]
tags: [sqlite, wal, gitops, problem-details]
deciders: [user, ai]
date: 2026-10-08
---

# 0035. メタデータ確認・編集 API および GitOps 逆同期アーキテクチャの採用 (0035-metadata-verification-and-editing-architecture.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-08

---

## 1. 背景と解決すべき課題 (Context & Problem)

1. **視覚的確認とメタデータ修正の摩擦**:
   - メンバーのペンライトカラー、順序、期生、ステータス、衣装画像（PhotoType）などのマスターデータは、テキストのテーブルや SQL 直接操作だけで確認・校正するのが極めて困難。
   - 実際の写真とペンライト発光色の組み合わせを人間が視覚的に確認（Human-in-the-Loop: [ADR-0024](./0024-photo-metadata-human-in-the-loop-architecture.md)）し、その場で修正を直接反映（Direct In-Place Mutation）できるバックエンド基盤が必要とされていた。
2. **マスターデータ同期と運用 DB 更新の二重管理リスク**:
   - [ADR-0021](./0021-gitops-master-data-synchronization-and-versioning-architecture.md) では `seeds/data/` を正本として SQLite WAL に同期する GitOps フローを採択している。
   - 一方で、運用 DB（SQLite）上で校正・変更した内容がコードリポジトリのシードファイルに還元できない場合、次回デプロイやシード同期時に手動修正内容が上書き破棄される危険性があった。
3. **過度なアップロード機能導入の回避（YAGNI原則）**:
   - 新規画像の収集や高解像度化はバッチスクリプト・公式 CDN 経由のオンデマンドプロキシ（[ADR-0033](./0033-on-demand-cache-through-image-proxy.md)）が担っており、Web UI からの multipart/form-data バイナリアップロードは設計スコープ外である。
   - 新規作成（Create）の拡張余地は残しつつ、現フェーズでは「メタデータの確認および更新（Update/Patch）」に特化した最小構成を確立する必要があった。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

1. **[DB] `members` テーブルへの `verified_at` カラム追加**:
   - `migrations/000003_add_member_verified_at.up.sql` において、`members` テーブルに `verified_at TEXT`（ISO 8601 UTC）カラムおよび検索インデックス `idx_members_verified` を配備する。
2. **[Contract] `Member` モデル拡張と管理用 DTO 定義**:
   - `pkg/model/member.go` の `Member` 構造体に `VerifiedAt *time.Time` フィールドを追加する。
   - `pkg/model/admin.go` に更新リクエスト DTO（`UpdateMemberPenlightRequest`, `UpdateMemberStatusRequest`, `SetPrimaryMemberImageRequest`, `UpdateMemberImagePhotoTypeRequest`）を定義し、外部依存ゼロのドメイン型として一元管理する ([ADR-0004](./0004-go-schema-as-single-source-of-truth.md))。
3. **[Backend] `model.Repository` によるメタデータ更新メソッドの集約**:
   - `pkg/model/repository.go` に以下のメソッドを定義し、`pkg/repository/sqlite.go` に SQLite WAL 実装を提供する ([ADR-0017](./0017-repository-interface-and-pure-go-sqlite-architecture.md))：
     - `GetMember(ctx, id)`: メンバー詳細および全登録写真（`member_images`）の取得
     - `UpdateMemberPenlight(ctx, id, penlight, markVerified)`: ペンライト色・順序の更新と自動 `verified_at` 記録
     - `UpdateMemberStatus(ctx, id, status, generation)`: ステータスおよび期生の更新
     - `MarkMemberVerified(ctx, id)`: スワイプ OK 用の確認完了日時更新
     - `SetPrimaryMemberImage(ctx, memberID, imageID)`: トランザクションによる代表写真切り替え
     - `UpdateMemberImagePhotoType(ctx, imageID, photoTypeID)`: 写真衣装種別の紐付け変更
4. **[Contract] RFC 9457 Problem Details 準拠の Admin REST API 提供**:
   - [ADR-0014](./0014-error-handling-and-minimal-problem-details.md) の 6 つの公式エラーコードに準拠し、`pkg/server/server.go` に以下のエンドポイントを実装する：
     - `GET /api/v1/admin/members/{id}`
     - `POST /api/v1/admin/members/{id}/verify`
     - `PATCH /api/v1/admin/members/{id}/penlight`
     - `PATCH /api/v1/admin/members/{id}/status`
     - `PUT /api/v1/admin/members/{id}/images/primary`
     - `PATCH /api/v1/admin/images/{id}/photo-type`
5. **[Backend] CORS プリフライトにおける PATCH / PUT メソッドの許可**:
   - クライアント SPA からのインライン更新を可能にするため、`corsMiddleware` の `Access-Control-Allow-Methods` に `PATCH` および `PUT` を追加する。
6. **[Tool] GitOps マスターデータへの逆同期エクスポートスクリプトの提供**:
   - `scripts/export_seeds.go` を提供し、SQLite 上の最新更新データを `seeds/data/members.json` に逆出力・フォーマット保存できるようにする。これにより Git コミットおよび PR 作成経由で GitOps 正本へ還元する循環フローを確立する。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### メタデータ更新・永続化方式の比較検討

| 評価項目 | 下書き・承認キュー方式 | 管理者専用別DB方式 | SQLite WAL 直接更新 ＋ GitOps逆同期 [採択] |
|---|---|---|---|
| **アーキテクチャの複雑さ** | 高（承認待ち状態・差分テーブル管理が必要） | 中（複数DBの運用保守・同期が必要） | **極小（単一 SQLite WAL ＋ DTO直更新）** |
| **トリアージ速度 (UX)** | 低（承認工程により即時反映されない） | 中 | **最高（キー操作・スワイプの瞬間に即時反映）** |
| **GitOps 整合性** | 複雑 | 困難 | **完全対応（`export_seeds.go` でシードへ還元）** |
| **外部依存性** | 甚大 | 甚大 | **ゼロ（標準ライブラリ＋Pure Go SQLite）** |

---

## 4. 得られる効果と運用規約 (Consequences & Enforcement)

### 4.1 ポジティブな影響 (Positive)
- **高速トリアージの実現**: ギャラリービューおよびスワイプビューから、ワンクリック/キーストロークで即座にデータが更新され、ストレスのない校正体験を提供。
- **データ不整合の防止**: `SetPrimaryMemberImage` のトランザクション保証により、代表写真が複数存在したりゼロになる不整合を機械的に排除。
- **GitOps の可逆的同期**: SQLite で人間が直感的に修正した成果を `export_seeds.go` で即座にコードリポジトリにコミット可能。

### 4.2 運用規約 (Enforcement)
- **直接 SQL 操作の禁止**: 運用時のメタデータ修正は、必ず本 ADR の Admin API または今後構築する UI を通じて実行し、サロゲートキー不整合やトリガー漏れを防ぐこと。
- **コンプライアンステストの維持**: 本 ADR の決定事項は `pkg/quiz/adr_compliance_test.go` の `TestADR0035_Compliance_MetadataEditing` で機械的に検証し、リグレッションを常時防止する。
