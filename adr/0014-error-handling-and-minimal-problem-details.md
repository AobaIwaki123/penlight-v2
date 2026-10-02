---
id: ADR-0014
title: エラー設計および最小 Problem Details 規約の採用
status: Accepted
scope: System
primary_category: APP
categories: [APP, DEV]
tags: [error-handling, rfc9457, problem-details, yagni]
deciders: [user, ai]
date: 2026-10-03
---

# 0014. エラー設計および最小 Problem Details 規約の採用 (0014-error-handling-and-minimal-problem-details.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-03

---

## 1. 背景と解決すべき課題 (Context & Problem)

Web API の設計において、エラーコードを「あったら丁寧」という動機で細分化しすぎると、以下の深刻な問題が発生します。
1. **YAGNI 違反と過剰な細切れ化**: `RES_GROUP_NOT_FOUND`、`RES_MEMBER_NOT_FOUND`、`RES_COLOR_NOT_FOUND` のようにリソースごとにコードを分けると、クライアントはどれも「データが見つかりません」と表示するだけであるにもかかわらず、コード定義が肥大化する。
2. **クライアント分岐の複雑化**: フロントエンド側で数十個のエラーコードをハンドリングする必要が生じ、メンテナンスコストが増大する。
3. **HTTP ステータスコードの形骸化**: HTTP 規格（RFC 9110）が定義するステータス（400, 401, 403, 404, 422, 500）の役割を無視し、オレオレコードで二重管理してしまう。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

RFC 9457（Problem Details for HTTP APIs）に準拠しつつ、**クライアントが実際に画面遷移やトースト表示などの挙動を変える必要がある最小限の 6 つのコードのみ** を採用する。

### 2.1 採用するエラーコード一覧（全 6 種）

| HTTP Status | エラーコード (`code`) | 発生条件 | クライアント（フロントエンド）の挙動・分岐 |
|---|---|---|---|
| **400** | `INVALID_PARAMS` | パラメータ形式不正、必須漏れ、画像サイズ超過 | 該当フォーム項目にインラインエラーを表示、またはトースト通知 |
| **401** | `UNAUTHORIZED` | 未ログイン、セッション切れ | ログインモーダルを表示、またはログイン画面へリダイレクト |
| **403** | `FORBIDDEN` | 一般ユーザーが `/admin` 系 API にアクセス | 「管理者権限がありません」警告画面を表示 |
| **404** | `NOT_FOUND` | 指定された ID（グループ、メンバー等）が存在しない | 「データが見つかりません」画面を表示、またはマスタ再同期 |
| **422** | `INSUFFICIENT_MEMBERS` | 絞り込み条件（期生等）の現役メンバーが 4 人未満で 4 択クイズを生成できない | **（ドメイン特有）** ユーザーに「全期生を対象にしてください」と条件緩和を提案 |
| **500** | `INTERNAL_ERROR` | サーバー内部の予期せぬ例外、DB 接続失敗 | 「サーバーで問題が発生しました」トーストを表示（自動で指数バックオフ 1 回再試行） |

### 2.2 レスポンス JSON 構造 (RFC 9457 最小構成)

```json
{
  "title": "Bad Request",
  "status": 400,
  "code": "INVALID_PARAMS",
  "detail": "Validation failed for request parameters",
  "invalid_params": [
    {
      "field": "member_id",
      "message": "Must be a valid TypeID starting with 'mem_'"
    }
  ]
}
```

- 詳細なバリデーションエラー（どの項目がどう悪いか）は個別コードに分けず、`invalid_params` 配列に集約する。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### エラー設計パターンの比較検討

| 方式 | 構造 | メリット | デメリット | 判定 |
|---|---|---|---|---|
| **HTTP ステータスのみ** | `{ "message": "Not Found" }` | 最もシンプル | 400 や 422 でクライアントがどの項目をどう直せばよいか機械的に判別不能 | **却下**: クライアント体験低下 |
| **過剰な独自コード乱立** | `RES_MEMBER_NOT_FOUND`, `IMG_TOO_LARGE` 等 20+ 個 | 一見細かく見える | YAGNI 違反。クライアント側の分岐が錯綜し、メンテナンス不能 | **却下**: AI slop・過剰設計 |
| **最小 Problem Details [採択]** | **標準 HTTP + 最小 6 コード + invalid_params** | **HTTP 規格に忠実、クライアント分岐が自明、実装 30 行** | ドメイン特有コードが増えた際は再評価が必要（現在は 1 つのみ） | **採用**: 最高のシンプルさと実用性 |

---

## 4. 効果とデメリット対策 (Consequences & Mitigations)

### プラスの効果
1. **Go バックエンド実装の最小化**: `pkg/apperror` は外部依存ゼロ、1 つの構造体と数行のヘルパー関数（`InvalidParams`, `NotFound`, `Unauthorized`）だけで 30 行以内で完結する。
2. **フロントエンド分岐の簡素化**: React/TypeScript 側の `switch (error.code)` ハンドラが数行で書け、テストや型ガードが極めて容易になる。

### デメリットと対策
- **懸念**: 今後機能追加で新しいドメイン固有エラーが必要になった場合どうするか。
- **対策**: 「クライアント側でそのコード固有の画面分岐（モーダル、特殊導線）が必要か？」を厳格に審査し、真に必要な場合のみ ADR を更新して追加する。
