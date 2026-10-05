---
name: doc-lifecycle
description: 調査・設計ノート（docs/notes/）でのユーザー壁打ちから、ユーザー認可に基づく意思決定記録（adr/）への昇格ライフサイクルを管理・統制する。
---

# ドキュメント・意思決定ライフサイクル規約 (doc-lifecycle)

> **管轄 ADR**: [ADR-0016](../../../adr/0016-adr-governance-and-immutable-sequential-architecture.md)

本スキルは、調査・設計の検討から確定した意思決定（ADR）への昇格フローを統制する。

## 2段階ライフサイクルの原則

AI エージェントは、いきなり ADR を起票してはならない。必ず以下の 2 段階を踏むこと。

```
[課題・技術調査] ──► docs/notes/XX_title.md で論点・トレードオフ整理 ──► ユーザーと壁打ち
                                                                           │
                                                                           ▼
[adr/README.md 一覧更新] ◄── adr/00XX-title.md 起票 ◄── 【ユーザーの明示的認可】
```

---

## フェーズ 1: 調査・壁打ち (`docs/notes/`)

新しい技術選定、アーキテクチャ設計、課題解決に取り組む際は、まず `docs/notes/` 配下に連番付きノートを作成する。

- **ファイル命名**: `docs/notes/XX_snake_case_topic.md`（例: `docs/notes/01_sqlite_migration_strategy.md`）
- **記載内容**:
  1. 背景と解決すべき課題 (Why)
  2. 比較検討した選択肢 (Options)
  3. 各選択肢のメリット・デメリット・トレードオフ
  4. 推奨案と論点
- **ユーザーとの対話**: ノートの要点をユーザーに提示し、方針について合意・フィードバックを得る。

---

## フェーズ 2: ADR への昇格と起票 (`adr/`)

方針が固まり、**ユーザーから「ADR を作成してください」「これで決定」等の明示的な認可を受けた場合のみ**、ADR を作成する。

### 1. ファイル命名と連番採番
- パス: `adr/00XX-kebab-case-title.md`
- 連番はグローバル連番（`0017`, `0018`, ...）。既存の最大番号 + 1 とする。

### 2. YAML Frontmatter の強制
ファイルの先頭に必ず以下のメタデータを記載する。

```yaml
---
id: ADR-00XX
title: 決定事項の簡潔なタイトル
status: Accepted
scope: System
primary_category: ARC # ARC, DOM, DAT, APP, DEV のいずれか
categories: [ARC, DEV]
tags: [go, k8s] # 公式 Allowlist から 3〜4 個選択
deciders: [user, ai]
date: YYYY-MM-DD
---
```

### 3. 公式タグ Allowlist の厳守
タグの表記揺れを防ぐため、`tags` には `adr/README.md` に定義された公式許可タグからのみ選択すること。アドホックなタグの追加は禁止。

- **ARC**: `go`, `k8s`, `cloudflare`, `gitops`, `monorepo`, `directory-structure`, `configuration`
- **DOM**: `typeid`, `uuidv7`, `single-source-of-truth`, `surrogate-key`, `extensible-group`
- **DAT**: `sqlite`, `wal`, `litestream`, `immutable-cache`, `s3-backup`
- **APP**: `local-first`, `pwa`, `quiz-strategy`, `oidc`, `problem-details`, `cielab`
- **DEV**: `toolchain`, `biome`, `knip`, `adr-governance`, `yagni`

### 4. `adr/README.md` の一覧更新
ADR を作成したら、必ず `adr/README.md` の「1. 意思決定一覧」テーブルに新しい連番とタイトルを追記する。

### 5. 具現化仕様のナンバリングとレイヤー接頭辞の強制 (Ref: ADR-0028)
ADR 本文の「2. 決定事項と具現化仕様」は、コンプライアンステスト（`ADR-00XX/Y`）による機械的トレーサビリティを担保するため、必ず連番（`1.`, `2.`, `3.`）で箇条書きし、先頭に対象レイヤー（`[Backend]`, `[Frontend]`, `[DB]`, `[Contract]` 等）を明記すること。

```markdown
## 2. 決定事項と具現化仕様 (Decision & Specification)
1. [Backend] ○○関数の提供
2. [Frontend] ○○コンポーネントにおける UI の提供
3. [Contract] ○○エラーレスポンスの返却
```

---

## 不変性の原則 (Append-only)

- 一度 `Accepted` となった ADR の本文直接改ざんは禁止。
- 過去の決定を変更・廃止する場合は、旧 ADR のステータスを `Superseded` に更新し、新しい ADR（`Superseded by ADR-00XX`）を起票して歴史的経緯を保全すること。
