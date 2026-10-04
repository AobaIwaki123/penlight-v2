---
name: k8s-gitops-guard
description: 自宅KubernetesクラスタにおけるArgoCD宣言的GitOps、Cloudflare Tunnel Ingress、単一Pod(Recreate)の安全なデプロイ・運用手順を統制する。
---

# Kubernetes デプロイ & ArgoCD GitOps 運用規約 (k8s-gitops-guard)

> **管轄 ADR**: [ADR-0003](../../../adr/0003-deployment-target-and-container-registry.md), [ADR-0012](../../../adr/0012-kubernetes-deployment-and-gitops-architecture.md)

本スキルは、自宅 Proxmox k8s クラスタ上で稼働する `penlight-v2` において、ArgoCD を用いた宣言的 GitOps、Cloudflare Tunnel Ingress による安全な外部公開、および SQLite WAL の単一ライター整合性を担保するための運用・検証手順を定める。

---

## 1. 絶対遵守ルール (Invariants)

1. **単一レプリカ (`replicas: 1`) & `Recreate` 戦略の強制**:
   - SQLite WAL の同時書き込みによるデータベース破損を構造的に防ぐため、Deployment は必ず `replicas: 1` かつ `strategy.type: Recreate` でなければならない（RollingUpdate による新旧 Pod の多重起動を禁止）。
2. **ArgoCD 宣言的 GitOps の徹底 (No Direct kubectl Apply)**:
   - 一時的なデバッグを除き、本番リソースを `kubectl edit` や個別の手動 apply で書き換えてはならない。すべてのマニフェスト変更は `deploy/` ディレクトリ配下にコミットし、GitOps 自動同期（`selfHeal: true`, `prune: true`）でクラスタへ反映させる。
3. **Cloudflare Tunnel によるポート開放ゼロの外部公開**:
   - ルーターのポート開放や固定 IP は使用せず、`ingressClassName: "cloudflare-tunnel"` および `cert-manager.io/cluster-issuer: letsencrypt-cloudflare` により `penlight.aooba.net` として安全に外部公開する。
4. **最小リソースクォータの設定**:
   - 単一バイナリの軽量性を活かし、CPU Request `20m` / Memory Request `32Mi`、CPU Limit `500m` / Memory Limit `128Mi` を維持してクラスタリソースを浪費しない。
5. **最小構成の環境変数 (ADR-0015 準拠)**:
   - マニフェスト内で設定する環境変数は `DATA_DIR: /data` を基本とし、不要な環境変数の追加・肥大化を禁止する。

---

## 2. デプロイ運用手順 (Deployment Runbook)

### 2.1. 初回クラスタ登録手順
ArgoCD にアプリケーションを初回登録する手順：

```bash
# 1. ArgoCD Application マニフェストの適用
kubectl apply -f deploy/argocd/app.yml

# 2. アプリケーション登録状態の確認
kubectl get application penlight -n argocd

# 3. 正常同期の確認 (SYNC: Synced, HEALTH: Healthy)
kubectl get application penlight -n argocd
```

### 2.2. 定常リリースフロー (GitOps)
1. **コード & マニフェスト変更**:
   - `git checkout -b <branch>` でブランチを作成し、機能実装やマニフェストを編集。
   - ローカル全検証を実施: `./scripts/verify-all.sh --stage`
2. **PR 作成 & マージ**:
   - GitHub 上で PR を作成し、CI がパスしたことを確認して `main` へマージ。
3. **自動ビルド & デプロイ同期**:
   - GitHub Actions (`.github/workflows/deploy.yml`) が起動し、GHCR へ最新コンテナを自動 push。
   - ArgoCD が `main` の変更を検知し、クラスタへ自動同期（ゼロダウンタイム反映）。

---

## 3. クラスタ状態・ヘルスチェックコマンド集

```bash
# 全リソース（Pod, Service, Ingress, PVC）の一括確認
kubectl get all,pvc,ingress -n penlight

# Pod のログ確認（リアルタイムログ）
kubectl logs -n penlight -l app=penlight -f

# Ingress / 証明書の状態確認
kubectl describe ingress penlight-ingress -n penlight

# ArgoCD アプリケーションの同期状態・差分確認
kubectl describe application penlight -n argocd
```

---

## 4. トラブルシューティング手順

| 事象 | 原因 | 対処方法 |
|---|---|---|
| `deploy: app path does not exist` | `main` ブランチに `deploy/` がまだマージされていない | PR を `main` にマージ後、ArgoCD をハードリフレッシュ (`kubectl annotate application penlight -n argocd argocd.argoproj.io/refresh=hard --overwrite`) する |
| `ErrImagePull / ImagePullBackOff` | GitHub Actions のコンテナビルド未完了、またはタグ名不一致 | `gh run list --workflow=Deploy` でビルド完了を確認。リポジトリが Public であることを確認 |
| `CrashLoopBackOff (permission denied)` | `/data` の所有権が非rootユーザーと不一致 | Dockerfile で `chown -R penlight:penlight /data` を指定し、PVC のパーミッションを確認 |
| `go: go.mod requires go >= X` | Dockerfile 内の Go ビルダーバージョン不足 | Dockerfile の `golang:<version>-alpine` を `go.mod` 以上（`container-guard` 参照）に更新する |

---

## 5. ローカル検証チェックリスト (マニフェスト変更時)

- [ ] `replicas: 1` かつ `strategy.type: Recreate` が設定されているか
- [ ] PVC のマウントパスが `/data`、`DATA_DIR` が `/data` になっているか
- [ ] Ingress の `ingressClassName` が `"cloudflare-tunnel"` になっているか
- [ ] リソース制限（Requests 32Mi / Limits 128Mi）が守られているか
- [ ] `actionlint` および `./scripts/verify-all.sh` が全てグリーンで通過したか
