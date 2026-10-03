---
name: github-actions-guard
description: GitHub Actionsの最新メジャーバージョン特定、サプライチェーン対策（コミットハッシュ固定）、aquaによるCLI固定、およびactionlintによるローカル静的検証を統制する。
---

# GitHub Actions 最新化・検証規約 (github-actions-guard)

> **管轄 ADR**: [ADR-0012](../../../adr/0012-kubernetes-deployment-and-gitops-architecture.md), [ADR-0013](../../../adr/0013-typescript-ai-agent-driven-development-toolchain.md)

本スキルは、GitHub Actions ワークフローを作成・更新する際に、AI の古い学習データによる「古い Action の採用」を防ぎ、サプライチェーン攻撃対策とローカルでの機械的検証を徹底するための運用手順を定める。

---

## 1. 絶対遵守ルール (Invariants)

1. **AI の学習データのみで Action バージョンを決めることの禁止**:
   - AI は過去のメジャーバージョン（例: checkout v4 等）を最新と誤認しやすいため、必ず `git ls-remote` を叩いて最新タグをリアルタイムに確認すること。
2. **フルコミットハッシュ（40桁 SHA-1）による固定**:
   - ミュータブルなタグ名（`@v7` や `@master`）での指定は禁止。サプライチェーン攻撃対策として、必ず 40 桁のコミットハッシュで指定し、末尾にバージョンコメントを添える。
   - 例: `uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1`
3. **CI 上でのバージョン未指定ツールの禁止 (aqua の利用)**:
   - `npm install -g <tool>` や `curl ... | sh` など、CI 実行時にバージョンが浮くインストールは禁止。必ず [`aqua.yaml`](../../../aqua.yaml) および [`aqua-checksums.json`](../../../aqua-checksums.json) でバージョンと SHA-256 チェックサムを固定する。
4. **ローカルでの `actionlint` 静的検証の必須化**:
   - push 前に必ずローカルで `actionlint` を実行し、YAML 構文、shellcheck、アクション入力値の妥当性を確認すること。
5. **CI 実行完了の待機禁止 (No Polling / No Watching)**:
   - ローカル検証（`actionlint` および `./scripts/verify-all.sh`）をパスして push / PR 作成した後は、リモート CI の完了を待つポーリング（`gh run watch` や `sleep` 等）を行わない。直ちに次の実装タスクまたはユーザー報告へ移行すること。

---

## 2. Action 最新バージョン & ハッシュ特定コマンドリスト

新しい Action を導入または更新する際は、以下のコマンドで最新タグとコミット SHA を取得する。

```bash
# 【最新タグ確認】v4, v5 に絞らず、末尾タグをすべて確認する（v6, v7 の見落とし防止）
git ls-remote --tags https://github.com/<owner>/<repo>.git | tail -n 15

# 【HEAD コミット確認】（タグが存在しない Action の場合）
git ls-remote https://github.com/<owner>/<repo>.git HEAD
```

---

## 3. aqua による CLI ツールのバージョン固定手順

CI で使用する CLI ツール（Biome, typos 等）を追加・更新する手順：

```bash
# 1. ツールを aqua.yaml に追加（バージョン固定）
aqua g <package_name> >> aqua.yaml

# 2. チェックサムファイルを更新（改ざん防止）
aqua update-checksum -a

# 3. 動作確認
aqua i -l
```

---

## 4. ローカル検証チェックリスト (変更時)

- [ ] `actionlint` がエラーゼロで通過したか
- [ ] すべての `uses:` が 40 桁のコミットハッシュでピン留めされているか
- [ ] 各コミットハッシュの末尾に `# vX.Y.Z` のコメントが付与されているか
- [ ] `./scripts/verify-all.sh` が全てグリーンで通過したか
