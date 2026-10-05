# 10. メタデータ確認・アノテーション（ギャラリー＆スワイプビュー）設計

## 1. 背景と目的 (Background & Objectives)

### 1.1 背景と課題
- **視覚的整合性の検証負荷**: メンバーのペンライトカラー（2色・順序）、期生・ステータス、衣装画像（PhotoType）などのメタデータは、テキストのテーブルやSQLだけで確認・校正するのが極めて困難。実際の写真とペンライト色の組み合わせを人間が見て初めて「色が微妙に違う」「衣装とタグがズレている」等の違和感に気づくことが多い。
- **管理運用の煩雑さ**: 個別ページを1画面ずつ開いてフォームを入力・保存する従来のCRUD管理画面では、全メンバーや特定期生の一括トリアージに膨大なクリック数と認知負荷がかかる。
- **実機プレビューの欠如**: クイズ本番で表示される「ペンライトカラーが発光したカード」の状態で確認できないため、色の再現性（ダークモードや背景色とのコントラスト）の評価が別工程になってしまう。

### 1.2 目的とアプローチ
1. **全体俯瞰（ギャラリービュー）**:
   - グループ・期生・ステータス単位でメンバーカードをグリッド表示し、画像・ペンライト発光色・名前を一覧で俯瞰。
   - ペンライト色が未設定、または画像が未登録などの異常・欠損データを視覚的にハイライトし、即座に修正対象を特定可能にする。
2. **高速トリアージ（マッチングアプリ風スワイプビュー）**:
   - マッチングアプリのカードスタックUIを採用。中央にメンバー画像と正解ペンライト発光プレビューを表示。
   - キーボード（←/→/↑）またはスワイプジェスチャーにより、「OK（確認完了）」「要修正（パレット/編集展開）」「スキップ（保留）」を高速にさばく。
3. **即時反映（Direct In-Place Mutation）**:
   - 少人数・個人による迅速なデータメンテナンスを重視し、修正内容は複雑な承認キューを介さず、API経由で即座に SQLite 本番データに反映する。

## 2. UI/UX コンセプト (UI Modes & User Flow)

### 2.1 画面全体の構造と共通コントロール
- **モード切り替え**: ヘッダーの SegmentedControl で「ギャラリービュー」と「スワイプ判定」をシームレスに切り替え可能。
- **フィルタ・スコープ**: グループ（日向坂46等）、期生（1期/2期/3期/4期等）、ステータス（現役/卒業等）、画像未登録・色未設定フィルタ。
- **進捗インジケータ**: 対象メンバー総数に対する確認進捗（例: `32 / 48 完了`）をリアルタイム表示。

### 2.2 Mode A: ギャラリービュー (Gallery View) の詳細体験
- **目的**: グループ全体の一覧性・バランス確認、特定メンバーへのピンポイントアクセス、衣装の網羅性チェック。
- **表示密度コントロール (Zoom & Density)**:
  - **Comfortable (大)**: 写真（3:4）、発光ペンライト、色名、期生、衣装フィルムストリップ、クイック編集ボタンを展開。
  - **Compact (中 - デフォルト)**: 写真＋左右色チップ＋名前。1画面に 15〜20 人が収まり、期生全体のバランスを確認しやすい。
  - **Overview (極小 - カラーマトリクス)**: ほぼサムネイルとペンライトカラーバーのみ。40人以上のグループ全体を一望し、色の偏りや画像抜けを瞬時に把握。
- **グルーピングと多軸ビュー**:
  - **メイン（メンバー軸）**: 1カード＝1メンバー。カード内に登録済み衣装（PhotoType）の小さなサムネイル列（フィルムストリップ）を内包し、衣装クリックで写真を切り替え・確認。期生単位（1期・2期…）でセクション分割。
  - **サブ（衣装軸 / PhotoType 俯瞰）**: 衣装ドロップダウンで「12th制服」等を選択すると、その衣装を着たメンバー一覧が並ぶ。未登録のメンバーは「未登録（グレー枠）」として表示され、収集漏れが一目瞭然。
- **メタデータ検証を加速するソート＆フィルタ**:
  - **色相順ソート（カラーグループ順）**: 期生順だけでなく「左手色順」で並び替え可能。同系色（パステルブルー等）を使うメンバーが隣同士に並ぶため、登録ミスや重複を一瞬で発見可能。
  - **特定色フィルタ**: カラーパレットから1色選ぶと、その色を含むメンバーのみハイライト/絞り込み。
- **ワンクリック・インラインアノテーション**:
  - **左右ワンクリック入替 (Swap 🔄)**: 左右の持ち手色が逆の場合、カード上のスワップボタン1クリックで即座に入替・DB保存。
  - **公式カラーパレット (Popover)**: 色チップをクリックすると、グループ公式色（14〜15色）のパレットが開き、ワンタップで即時更新。
  - **Primary画像切替**: 衣装サムネイルのピン留めアイコンで、クイズ出題時のデフォルト画像を即時変更。

### 2.3 Mode B: スワイプ / カードビュー (Swipe Annotation View)
- **目的**: 1メンバーに集中した高速チェック、実機クイズカードの忠実な再現による視覚的検証。
- **カードビジュアル（マッチングアプリ風）**:
  - 画面中央にスタック配置される大型カード（背面に次のカードがチラ見えする演出）。
  - クイズ本番画面と同様のネオングローペンライトが左右に配置され、暗背景でリアルな発光感を再現。
  - メンバーのプロフィール情報（名前・かな・期生・ステータス）。
- **操作アクションとキーバインド**:
  - **右スワイプ / `→` / `D`**: **OK（確認完了）** ── 緑のアニメーションと共にカードが右へ飛び、次へ進む。
  - **左スワイプ / `←` / `A`**: **要修正** ── 赤/黄のアニメーションと共に編集パネル（カラーピッカー / 画像選択）が展開。
  - **上スワイプ / `↑` / `W`**: **スキップ** ── 保留として上へ飛び、キューの末尾へ再送またはスキップ。
  - **インライン即時修正 (`E` キー)**: スワイプせずその場でペンライト色をドロップダウン/パレットから選択し、確定と同時に次へ送る。
  - **アンドゥ (`Z` キー / 戻るボタン)**: 直前のスワイプ判定を取り消してカードを戻す。

## 3. アノテーション対象とデータモデル設計 (Data & Domain Schema)

### 3.1 アノテーション対象属性
| エンティティ | 属性 | 型 | 説明・編集UX |
|---|---|---|---|
| `members` | `penlight_left_color_id` | `ID (col_...)` | 左手ペンライト色。公式パレットから選択 |
| `members` | `penlight_right_color_id` | `ID (col_...)` | 右手ペンライト色。公式パレットから選択 |
| `members` | `penlight_ordered` | `BOOLEAN` | 左右順序の有無（反転許容か固定か） |
| `members` | `status` | `TEXT` | `active` / `graduated` / `hiatus` |
| `members` | `generation` | `INTEGER` | 期生番号（1, 2, 3...） |
| `member_images` | `photo_type_id` | `ID (pht_...)` | 写真の衣装・用途カテゴリの変更・紐付け |
| `member_images` | `is_primary` | `BOOLEAN` | デフォルト代表画像の切り替え |

### 3.2 確認状態（Verification Status）の永続化
即時更新に加え、スワイプやギャラリーで「確認済み」かどうかを判定・フィルタできるように、以下のカラム拡張またはメタデータ管理を検討する。

- **推奨案: `members` テーブルへの `verified_at` カラム追加 (DAT / ADR-0004)**
  - `verified_at DATETIME NULL`: 人間がOK判定、または編集確定した時刻を記録。
  - メリット:
    - 「未確認のみスワイプ」「未確認メンバーのみギャラリー表示」がSQLレベルで極めて高速に実行可能。
    - マスターデータ更新や新期生追加時に `verified_at = NULL` とすることで、自動的に再トリアージ対象となる。
    - 差分履歴テーブル等を持たせない最小構成（KISS原則）を維持。

### 3.3 Go モデル層およびリポジトリ層の拡張 ([ADR-0017](../../adr/0017-repository-interface-and-pure-go-sqlite-architecture.md))
`pkg/model/repository.go` に以下のミューテーションメソッドを追加する。

```go
// Metadata Annotation Mutations
UpdateMemberPenlight(ctx context.Context, id ID, penlight PenlightPair) error
UpdateMemberStatus(ctx context.Context, id ID, status MemberStatus) error
MarkMemberVerified(ctx context.Context, id ID) error
SetPrimaryMemberImage(ctx context.Context, memberID ID, imageID ID) error
UpdateMemberImagePhotoType(ctx context.Context, imageID ID, photoTypeID ID) error
```

## 4. API およびアーキテクチャ境界 (API & Responsibility Boundaries)

### 4.1 エンドポイント設計 (Minimal REST Endpoints)
[ADR-0014](../../adr/0014-error-handling-and-minimal-problem-details.md) の RFC 7807 規約に準拠し、最小構成のエンドポイントを定義。

| メソッド | パス | ペイロード / パラメータ | 処理概要 |
|---|---|---|---|
| `GET` | `/api/v1/master` | クエリなし | 既存拡張。メンバーに `verified_at` を含めて返却（オフラインキャッシュ対応） |
| `POST` | `/api/v1/admin/members/{id}/verify` | なし | メンバーを「確認済み」としてマーク（スワイプ OK アクション） |
| `PATCH` | `/api/v1/admin/members/{id}/penlight` | `{ "left_color_id": "...", "right_color_id": "...", "ordered": true }` | ペンライト色・順序の即時更新（自動で verified_at も更新） |
| `PATCH` | `/api/v1/admin/members/{id}/status` | `{ "status": "active" }` | ステータス（現役/卒業等）の更新 |
| `PUT` | `/api/v1/admin/members/{id}/images/primary` | `{ "image_id": "img_..." }` | 代表画像の指定 |
| `PATCH` | `/api/v1/admin/images/{id}/photo-type` | `{ "photo_type_id": "pht_..." }` | 画像の衣装カテゴリタグ付け |

### 4.2 エラーハンドリング ([ADR-0014](../../adr/0014-error-handling-and-minimal-problem-details.md))
- 存在しないメンバー・画像・色ID指定時: `404 Not Found` (`urn:penlight:error:not_found`)
- 不正なリクエストボディ（必須欠落・無効な色IDフォーマット）: `400 Bad Request` (`urn:penlight:error:bad_request`)
- システムエラー: `500 Internal Server Error` (`urn:penlight:error:internal_error`)

### 4.3 フロントエンド構成とルーティング ([ADR-0011](../../adr/0011-directory-structure-and-responsibility-boundaries.md))
- パス: `/annotate`
- 完全静的エクスポート (`output: 'export'`) を維持し、Go 単一バイナリに内包。
- クライアント状態（SWR / React State）で楽観的更新（Optimistic UI）を行い、スワイプやクリックの瞬間にUIへ反映しつつバックグラウンドで即時 API 送信。

## 5. 実装ロードマップと段階的検証 (Phased Implementation & Verification)

### Phase 1: モデル・DB・リポジトリ層の拡張 (Contract & Backend)
1. `pkg/model/member.go` に `VerifiedAt *time.Time` フィールド追加
2. `migrations/` に SQLite カラム追加 DDL を適用 (`ALTER TABLE members ADD COLUMN verified_at DATETIME;`)
3. `pkg/model/repository.go` に更新メソッドを定義し、`pkg/repository/` に SQLite WAL 実装
4. `./scripts/generate-all.sh` で TS 型定義・スキーマ・ER図を自動同期
5. バックエンド API ハンドラー（`cmd/server/`）の実装と単体テスト（`go test ./...`）

### Phase 2: ギャラリービューの実装 (Frontend Gallery)
1. `/annotate` ページの土台作成（Mantine `SimpleGrid`, ヘッダーツールバー）
2. 3段階ズーム表示（Comfortable / Compact / Overview）
3. 期生別グルーピング＆衣装フィルムストリップ表示
4. 左右カラーのワンクリック入替 (Swap 🔄) & 公式カラーポップオーバーピッカー
5. 色相順ソート・特定色フィルタの実装

### Phase 3: スワイプ / カードビューの実装 (Frontend Swipe & Triage)
1. Tinder 風カードスタックコンポーネント（Framer Motion または CSS Transform）
2. ネオングロー発光ペンライトのプレビュー表示
3. キーボードショートカット (`A`: 要修正, `D`: OK, `W`: スキップ, `E`: 即時色変更, `Z`: Undo)
4. スワイプジェスチャー連動と即時 API コール



