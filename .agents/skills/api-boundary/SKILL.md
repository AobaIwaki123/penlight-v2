---
name: api-boundary
description: 4つの最小環境変数と6つの公式エラーコード（Problem Details）を強制し、設定散乱と過剰なエラー細分化を防止する。
---

# API 境界・最小構成規約 (api-boundary)

> **管轄 ADR**: [ADR-0014](../../../adr/0014-error-handling-and-minimal-problem-details.md), [ADR-0015](../../../adr/0015-minimal-configuration-and-secrets-management.md)

本スキルは、最小環境変数アーキテクチャ (ADR-0015) および最小 Problem Details 規約 (ADR-0014) を統制する。

## 1. 最小環境変数仕様 (ADR-0015)

外部注入が必要な機密値とデータパスの **4 変数のみ** に限定する。ポート番号やタイムアウト等の内部定数を安易に環境変数化してはならない。

| 環境変数名 | 本番必須 | デフォルト値 | 説明 |
|---|---|---|---|
| `DATA_DIR` | 任意 | `./data` | SQLite および画像アセットの格納先パス |
| `SESSION_SECRET` | **必須** | (ローカル固定値) | Cookie 署名用シークレットキー (32バイト以上) |
| `GOOGLE_CLIENT_ID` | 任意 | 空 (ゲストモード) | Google OIDC クライアント ID |
| `GOOGLE_CLIENT_SECRET` | 任意 | 空 (ゲストモード) | Google OIDC クライアントシークレット |

- **定数固定するもの**: `PORT=8080`, `DB_BUSY_TIMEOUT=5s`, `MAX_UPLOAD_SIZE=10MB` 等はすべてコード内 `const` に固定する。

---

## 2. 最小エラーコード規約 (ADR-0014)

クライアント側の画面分岐に必要な **以下の 6 コードのみ** を使用する。独自エラーコードの追加は禁止。

| HTTP Status | エラーコード (`code`) | 発生条件 | クライアント側の挙動 |
|---|---|---|---|
| **400** | `INVALID_PARAMS` | パラメータ不正・サイズ超過 | フォーム項目にエラー表示 |
| **401** | `UNAUTHORIZED` | 未ログイン・セッション切れ | ログインモーダル表示 |
| **403** | `FORBIDDEN` | 管理者権限不足 | 権限エラー警告画面表示 |
| **404** | `NOT_FOUND` | 指定 ID が存在しない | 404 画面表示 |
| **422** | `INSUFFICIENT_MEMBERS` | 4択クイズ生成不能 (メンバー不足) | 対象期生の緩和を提案 |
| **500** | `INTERNAL_ERROR` | サーバー内部例外・DB接続失敗 | エラートースト表示（再試行） |

- **レスポンス形式**: RFC 9457 Problem Details (`title`, `status`, `code`, `detail`, `invalid_params[]`) に統一する。
- バリデーションの詳細は新規エラーコードを作らず、`invalid_params` 配列に集約すること。
