---
id: ADR-0015
title: 最小環境変数および機密性分離アーキテクチャの採用
status: Accepted
scope: System
primary_category: ARC
categories: [ARC, DEV]
tags: [configuration, environment-variables, twelve-factor, secrets, yagni]
deciders: [user, ai]
date: 2026-10-03
---

# 0015. 最小環境変数および機密性分離アーキテクチャの採用 (0015-minimal-configuration-and-secrets-management.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-03

---

## 1. 背景と解決すべき課題 (Context & Problem)

Twelve-Factor App の原則を「設定可能な項目をすべて環境変数化すべき」と曲解すると、ポート番号、タイムアウト値、バッファサイズ、ログ形式、CORS 設定など十数個以上の環境変数が乱立します。
これには以下の問題があります。
1. **設定の散乱と脆弱性**: 開発者がローカルで起動するために長大な `.env` を記述する必要が生じ、オンボーディングの摩擦となる。
2. **YAGNI 違反**: 運用上変更される可能性が極めて低い内部定数（SQLite のビジータイムアウト 5 秒、画像上限 10MB、待受ポート 8080 等）まで外部化され、認知負荷が増大する。
3. **k8s マニフェストの肥大化**: 大量の ConfigMap と Secret が必要になり、IaC の保守コストが増大する。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

環境変数を **「外部から注入しなければ動作しない機密値（Secrets）」と「k8s PVC のマウントパス」の 4 つのみに極小化** する。

### 2.1 採用する環境変数（全 4 種）

| 環境変数名 | デフォルト値 (Local) | 本番設定 (k8s) | 必須 | 注入が必要な理由 |
|---|---|---|---|---|
| `DATA_DIR` | `./data` | `/data` | 任意 | k8s PVC マウント先のベースディレクトリ（DB とアセットを一元管理） |
| `SESSION_SECRET` | `dev-insecure-secret...` | ランダム 64 hex 文字列 | **本番必須** | Git にコミットできない Cookie 署名用秘密鍵 (>= 32 bytes) |
| `GOOGLE_CLIENT_ID` | (空文字: ゲストモード) | Google 発行 ID | OAuth 時 | 外部連携 OAuth クライアント ID |
| `GOOGLE_CLIENT_SECRET` | (空文字: ゲストモード) | Google 発行 シークレット | OAuth 時 | 外部連携 OAuth 秘密鍵 |

### 2.2 コード内定数 (`const`) に固定するもの
- `PORT = 8080`: ポート変更は k8s Service や Docker ポートフォワード側で解決。
- `DB_BUSY_TIMEOUT = 5 * time.Second`: WAL 最適値。
- `MAX_UPLOAD_SIZE = 10 * 1024 * 1024` (10MB): メンバー写真上限。
- `SESSION_TTL = 30 * 24 * time.Hour` (30日)。
- `CORS`: Go embed.FS から同一オリジン配信するため本番は不要。ローカルは localhost をコード内許可。
- `LOG_FORMAT`: 本番コンテナは JSON、開発時は Text を自動判別。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### 設定管理方式の比較検討

| 方式 | 構造 | メリット | デメリット | 判定 |
|---|---|---|---|---|
| **全項目環境変数化** | 18+ 個の環境変数 | あらゆる値を外部から変更可能 | YAGNI 違反。`.env` 管理が破綻し、設定漏れ事故が増加 | **却下**: 過剰設計 |
| **設定ファイル (YAML/TOML)** | `config.yaml` 読み込み | 階層構造で設定可能 | コンテナ内へのファイル配置が必要で k8s 運用と不整合 | **却下**: クラウドネイティブ不適合 |
| **最小環境変数＋定数固定 [採択]** | **機密値＋データパスの 4 変数のみ** | **ゼロコンフィグ起動可能、Go 実装 30 行、ConfigMap 不要** | 内部定数を変更する際は再ビルドが必要（頻度は極小） | **採用**: 最高の開発体験と安全運用 |

---

## 4. 効果とデメリット対策 (Consequences & Mitigations)

### プラスの効果
1. **ゼロコンフィグローカル起動**: 開発者はリポジトリを clone して `go run cmd/server/main.go` を叩くだけで、`.env` なしで即座に `./data` 配下に DB とアセットフォルダが作られ動作する。
2. **Go コードのシンプル化**: 外部ライブラリ（viper や envconfig 等）を使わず、標準 `os.Getenv` だけで `pkg/config` が 30 行以内で完結する。
3. **k8s の極小化**: ConfigMap が全廃され、Secret 1 つを Pod に注入するだけの最小マニフェストになる。

### デメリットと対策
- **懸念**: ポート番号を変更したくなった場合。
- **対策**: コンテナ内ポートは 8080 固定とし、ホスト側や k8s Service の `targetPort: 8080` / `port: 80` のマッピング層で吸収する。
