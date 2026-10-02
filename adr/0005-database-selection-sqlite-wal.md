# 0005. データベースとしての SQLite (WAL モード) の採用 (0005-database-selection-sqlite-wal.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-03

---

## 1. 背景と解決すべき課題 (Context & Problem)

旧システムは Google BigQuery に直結していたため、1 回のクエリで数秒の遅延が発生し、月額コストやローカル開発の複雑さが課題でした。
新システム（`penlight-v2`）では、アイドルマスタデータ（グループ・カラー・メンバー）の参照だけでなく、**ユーザーごとの正答・誤答ログ（AnswerLog）の逐次書き込みや苦手分析** を、外部 DB サーバーなしで軽量・安全に実現する必要があります。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

データベースとして **Pure Go SQLite (`modernc.org/sqlite`) + WAL (Write-Ahead Logging) モード** を採用し、k8s PVC 上で永続化する。

1. **ジャーナルモード**: `PRAGMA journal_mode = WAL;` を起動時に強制。
2. **同時実行モデル**: 「無制限の並行読み取り（SELECT）」と「単一スレッドの書き込み（INSERT/UPDATE）」がロック競合せず完全並行で動作。
3. **耐久性担保**: k8s の PersistentVolumeClaim (5Gi) 上に `penlight.db` を配置し、**Litestream** によるオブジェクトストレージ（Cloudflare R2 / MinIO）への秒単位リアルタイム差分バックアップを標準化。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### 2026年時点の組み込み・軽量データベース比較検討

| データベース | 内部モデル | 特性・強み | 本システムへの適合性 | 判定 |
|---|---|---|---|---|
| **DuckDB** | 列指向 OLAP | 数千万件の集計・Parquet 解析が極速 | 不適。単一行の更新・逐次回答 INSERT (OLTP) には向かない | **見送り** |
| **libSQL (Turso)** | SQLite 分散フォーク | HTTP レプリケーション、エッジ同期 | 不要。自宅 k8s 単一 Pod で完結するため分散オーバーヘッドが無駄 | **見送り** |
| **PGlite** | WASM 組み込み Postgres | ブラウザ/Node.js で Postgres が動く | 不適。Go バックエンドからの組み込み運用には成熟度で劣る | **見送り** |
| **PostgreSQL (CloudNativePG)** | クラスタ型 RDBMS | 複数 Pod からの同時書き込み、高度な JSON | 将来用。単一 Pod で済む現フェーズではリソース浪費 (数百MB) | **将来候補** |
| **SQLite WAL [採択]** | **行指向 OLTP 組み込み** | **ゼロコンフィグ、Cgo不要、WALで並行読み書き** | **最適。クイズ出題と回答ログ追記に最も壊れにくく軽量** | **採用** |

- **決定打**: SQLite の WAL モードは秒間数千トランザクションの INSERT を余裕で処理可能。ファン向けクイズアプリの書き込み負荷（秒間数十〜数百件）に対して十分すぎる性能を持ち、サーバープロセス死・ネットワーク分断のリスクがゼロ。

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響 (Positive)
- **運用負荷ゼロ**: 外部 DB サーバーの死活監視・バージョンアップ・接続プール管理が不要。
- **超軽量フットプリント**: メモリ消費は数 MB。ローカル開発も `go run` だけで即座に SQLite ファイルが初期化される。
- **高可用バックアップ**: Litestream により、ハードウェア障害時も直前（数秒前）の状態まで自動復元可能。

### 留意点と対策 (Negative & Mitigation)
- **複数 Pod からの同時書き込み制限**:
  - SQLite は単一ライター制約があるため、Deployment は `replicas: 1` かつ `Recreate` 戦略で運用する。
  - 将来的にアクセスが急増し、複数 Pod への水平スケールアウトが必要になった段階で PostgreSQL（CloudNativePG）への移行を検討する。
