# 0012. Kubernetes デプロイおよび ArgoCD GitOps アーキテクチャの採用 (0012-kubernetes-deployment-and-gitops-architecture.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-03

---

## 1. 背景と解決すべき課題 (Context & Problem)

既存の別リポジトリ（`/Users/aobaiwaki/journee/k8s`）では、GCP Container Registry (GCR) の非推奨化やサービスアカウント JSON 秘密鍵（`key.json`）の失効トラブル、重量級 Node.js ランタイム（3 レプリカ・数百 MB メモリ）によるリソース圧迫が課題となっていました。
新システム（`penlight-v2`）では、Go 単一バイナリおよび SQLite WAL 永続化に最適化した、**極限まで軽量で壊れない k8s デプロイ・GitOps パイプライン** を構築する必要があります。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

自宅 Proxmox k8s クラスタにおいて、以下のデプロイ・GitOps 構成を正式採用する。

1. **Cloudflare Tunnel Ingress**: `ingressClassName: "cloudflare-tunnel"` により、ルーターのポート開放や DDNS 不要で `penlight.aooba.net` として安全に外部公開。
2. **cert-manager**: `letsencrypt-cloudflare` による Let's Encrypt 自動 TLS 証明書管理。
3. **ArgoCD 宣言的 GitOps**: `deploy/overlays/prod` を監視し、マニフェスト変更をゼロダウンタイム自動同期。
4. **単一 Pod & PVC (Recreate 戦略)**:
   - SQLite WAL の単一ライター整合性を担保するため、`replicas: 1` かつ `strategy.type: Recreate` を強制（多重マウントによる DB 破損を防止）。
   - PersistentVolumeClaim: 5Gi (Longhorn / local-path)。
   - リソースクォータ: CPU `20m` / Memory `32Mi` (Request), CPU `500m` / Memory `128Mi` (Limit)。
5. **GitHub Packages (`ghcr.io`)**: GCR およびサービスアカウント JSON 秘密鍵を全廃し、GitHub Actions トークンと直接連携。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### インフラ運用方式の比較検討

| 項目 | 従来 (`journee/k8s`) | 本システム (`penlight-v2`) [採択] | 選定理由・トレードオフ |
|---|---|---|---|
| **コンテナレジストリ** | GCR (`gcr.io`) + JSON 秘密鍵 | **GitHub Packages (`ghcr.io`)** | GCR 廃止リスク・秘密鍵失効事故を完全排除 |
| **レプリカ数・戦略** | 3 レプリカ (RollingUpdate) | **1 レプリカ (Recreate)** | SQLite の単一ライター制約を守りデータ破損を防止 |
| **リソース消費** | メモリ 300〜600MiB | **メモリ 15〜32MiB (1/10以下)** | 自宅 k8s の空きリソースを大幅に節約 |
| **外部公開** | Cloudflare Tunnel | **Cloudflare Tunnel (継承)** | 固定 IP・ポート開放不要の最高セキュリティ |

- **決定打**: 本システムは Go 単一バイナリ（15MB）で数千 QPS を処理可能なため、複数レプリカによる分散は不要。Recreate 戦略を採用することで、ローリングアップデート時に新旧 Pod が同時に同一 SQLite ファイルに書き込もうとする「ロック競合・DB 破損事故」を構造的に排除できる。

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響 (Positive)
- **運用維持費・鍵管理コストゼロ**: クラウド請求ゼロ、定期的な認証鍵ローテーション業務が消滅。
- **高密度クラスタ稼働**: メモリ 32MiB で動くため、自宅サーバの他ワークロードに干渉しない。
- **完全自動化 GitOps**: GitHub へのコミット・マージのみで ArgoCD が自動反映。

### 留意点と対策 (Negative & Mitigation)
- **デプロイ時の一時ダウンタイム (数秒)**:
  - `Recreate` 戦略のため、旧 Pod 停止から新 Pod 起動までの間に約 3〜5 秒の切断が発生する。
  - 対策: クライアント側が PWA (Local-First) で動作しているため、ユーザーの画面上では通信断が一切感知されず、オフラインキューイングされるため実質無停止となる。
