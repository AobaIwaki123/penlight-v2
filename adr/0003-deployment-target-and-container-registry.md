# 0003. デプロイ環境およびコンテナレジストリ選定 (0003-deployment-target-and-container-registry.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-03

---

## 1. 背景と解決すべき課題 (Context & Problem)

旧システムは Google Cloud Platform (GCP) 上の Google Container Registry (GCR) および Cloud Run に依存していました。
しかし、以下の課題が生じていました。
1. **クラウドランニングコスト**: 個人ファンアプリでありながら、マネージドサービスの月額コストが発生。
2. **プラットフォーム廃止リスク**: GCR の非推奨化・Artifact Registry への強制移行など、外部仕様変更に伴うメンテナンス負債。
3. **認証鍵管理の負荷**: GCP サービスアカウントの JSON 秘密鍵（`key.json`）の失効やローテーション管理が運用リスクとなる。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

1. **デプロイ環境**: **自宅 Proxmox k8s クラスタ** 上で稼働させ、**Cloudflare Tunnel / Ingress** 経由で `penlight.aooba.net` として外部公開する。
2. **コンテナレジストリ**: **GitHub Packages (`ghcr.io`)** を採用し、GitHub Actions CI/CD パイプラインと直結する。
3. **コンテナ形態**: Go 1.16+ `embed.FS` によりフロントエンド静的ビルドを内包した **単一バイナリ Pod (Single Pod)** で運用する。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### インフラ・レジストリ構成の比較検討

| 構成案 | 月額コスト | 外部依存・廃止リスク | 運用・秘密鍵管理 | 判定 |
|---|---|---|---|---|
| **A: GCP (Cloud Run + GCR)** | 有料 (従量課金) | 高い (GCR廃止等) | 煩雑 (GCP JSON秘密鍵必須) | **却下**: コスト・負債 |
| **B: 自宅 k8s + Docker Hub** | 無料 | 低い | 普通 (Docker Hub レート制限) | **見送り**: Pull制限リスク |
| **C: 自宅 k8s + ghcr.io [採択]** | **完全無料** | **極めて低い** | **極めて容易 (GitHub Token直結)** | **採用**: 既存資産活用・コストゼロ |

- **決定打**: 自宅 Proxmox クラスタに K3s/Talos 基盤が既に稼働しており、Cloudflare Tunnel を使うことで固定 IP やルーターのポート開放なしで安全に外部公開可能。さらに GitHub Packages (`ghcr.io`) は GitHub Actions と認証トークンが自動連携するため、秘密鍵の漏洩・失効リスクがゼロになる。

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響 (Positive)
- **インフラ費用ゼロ**: クラウド維持費を完全撤廃。
- **構成の極小化**: 単一 Pod（15〜30MB コンテナ）で完結するため、k8s マニフェストやネットワーク設定の障害点が最小限。
- **安全な外部公開**: Cloudflare の DDoS 防御、自動 SSL/TLS 終端（Let's Encrypt）、WAF 保護を無料で享受。

### 留意点と対策 (Negative & Mitigation)
- **自宅環境の停止リスク (停電・回線障害)**:
  - 対策: Litestream による外部ストレージ（Cloudflare R2 / MinIO）へのリアルタイム差分バックアップを標準化し、障害時も他環境へ即時リストア可能とする。
