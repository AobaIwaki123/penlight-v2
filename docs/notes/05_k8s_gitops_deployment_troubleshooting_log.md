# Kubernetes デプロイ トラブルシューティング＆デバッグログ

本ドキュメントは、Proxmox 上の自作 Kubernetes クラスタおよび ArgoCD への初回デプロイ時に発生した考慮漏れ・事象・原因・対応ログを記録したものである。

---

## 1. 発生事象と対応ログ

### 事象 1: Ceph RBD PVC への非 root 書き込み権限不足
- **エラー**:
  ```text
  failed to create server: failed to initialize repository: failed to open sqlite database: open /data/penlight.db: permission denied
  ```
- **原因**:
  - Dockerfile ではセキュリティ担保のため非 root ユーザー（`UID: 10001`, `GID: 10001`）で実行している。
  - Kubernetes の Ceph CSI (RBD) で動的プロビジョニングされた Volume はデフォルトで `root:root (0:0)` 権限でマウントされるため、コンテナ起動時に `/data` 配下に SQLite ファイルを作成できず Permission Denied となった。
- **対処**:
  - `deploy/deployment.yaml` の Pod `spec.securityContext` に以下を追加。
    ```yaml
    securityContext:
      fsGroup: 10001
      runAsUser: 10001
      runAsGroup: 10001
    ```
  - Kubernetes がボリュームマウント時に `/data` の所有グループを GID 10001 に再帰設定することで解消。

---

### 事象 2: 本番環境における `SESSION_SECRET` 未注入による Fail-Fast
- **エラー**:
  ```text
  failed to load config: SESSION_SECRET is required (must be at least 32 bytes)
  ```
- **原因**:
  - [ADR-0015](file:///Users/aobaiwaki/penlight-v2/adr/0015-minimal-configuration-and-secrets-management.md) により、セキュアな Cookie 署名を保証するため、ローカル開発環境以外の本番環境では `SESSION_SECRET`（32バイト以上）が未設定の場合に即座にフェイルファストする設計となっていた。
  - 初期デプロイマニフェストにおいて、Secret からの環境変数注入設定が漏れていた。
- **対処**:
  - クラスタ内に Kubernetes Secret `penlight-secret` を作成。
  - `deploy/deployment.yaml` の `env` に以下を追加。
    ```yaml
    - name: SESSION_SECRET
      valueFrom:
        secretKeyRef:
          name: penlight-secret
          key: SESSION_SECRET
    ```

---

### 事象 3: ランタイムコンテナにおける SQL ファイル未配置による起動失敗
- **エラー**:
  ```text
  failed to create server: failed to read migration SQL: file not found in search paths: migrations/000001_init.up.sql
  ```
- **原因**:
  - `pkg/server/server.go` の初期実装では、マイグレーション SQL (`migrations/000001_init.up.sql`) やシードデータ (`seeds/seed.sql`) をカレントディレクトリおよび親ディレクトリからの相対パス探索 (`os.ReadFile`) で取得していた。
  - マルチステージ Dockerfile の runner ステージでは、バイナリ `/app/server` のみがコピーされており、コンテナ内ファイルシステムに SQL ファイルが存在しなかった。
- **対処**:
  - ディスクファイルへの実行時依存を完全に排除し、[ADR-0002](file:///Users/aobaiwaki/penlight-v2/adr/0002-backend-go-architecture.md)（単一バイナリ）の理念を徹底。
  - `migrations/embed.go` および `seeds/embed.go` を新設し、Go の `embed.FS` でコンパイル時にバイナリ内に完全内包した（設計詳細は ADR-0025 を参照）。

---

### 事象 4: ArgoCD におけるコンテナイメージ更新の自動検知
- **課題**:
  - GitHub Actions でイメージをビルド＆プッシュしても、マニフェスト側のイメージタグが `:latest` のままだと、ArgoCD から見て Git リポジトリに差分が生じず、Pod の自動 Recreate / ImagePull がトリガーされない。
- **対処**:
  - `.github/workflows/deploy.yml` において、ビルドしたイメージタグを短縮コミットハッシュ（`${GITHUB_SHA::7}`）としてプッシュ。
  - ワークフロー内で `deploy/deployment.yaml` の `image` フィールドを当該コミットハッシュで書き換え、`[skip ci]` を付与してリポジトリに自動コミット・プッシュする仕組みを導入。
  - ArgoCD がリポジトリの更新差分を検知し、宣言的かつ自動的に最新イメージで Pod を Recreate する GitOps サイクルを確立。
