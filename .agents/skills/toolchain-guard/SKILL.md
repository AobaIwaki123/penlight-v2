---
name: toolchain-guard
description: 超高速機械的ツール群（Biome, typos, Knip）によるミリ秒自動検証・コード整形・デッドコード排除・自律修復ループを統制する。
---

# AI 駆動開発ツールチェーン・自律修復規約 (toolchain-guard)

> **管轄 ADR**: [ADR-0013](../../../adr/0013-typescript-ai-agent-driven-development-toolchain.md)

本スキルは、AI エージェントおよび開発者がコードを生成・編集した直後に実行すべき「ミリ秒自律検証・自動修復パイプライン」を統制する。人間へのレビュー提出前に、機械的ツールで誤字脱字、構文崩れ、デッドコードを根絶する。

---

## 1. 高速自律修復パイプライン (Self-Healing Loop)

コード作成・変更後は、以下の順序でコマンドを実行し、エラーを自己修正してからユーザーに結果を報告すること。

```
[コード編集・生成]
       │
       ▼
1. Biome Check (15ms)    ─── 警告/違反 ──► `biome check --write` で自動整形・修復
       │ 合格
       ▼
2. typos (20ms)          ─── 誤字検出  ──► `typos -w` で自動修正
       │ 合格
       ▼
3. tsc (型検査)          ─── 型エラー  ──► エラー箇所を自己修正
       │ 合格
       ▼
4. Knip (衛生監査)       ─── ゾンビ検出 ──► `npx knip --fix` で不要 export/孤立依存を削除
       │ 合格
       ▼
[作業完了・ユーザーへの報告]
```

---

## 2. 日常コマンドリスト (Quick Commands)

### 一括自動検証・フォーマット
```bash
# 【最速】構文・リント・インポート順序のミリ秒自動修正 (Biome)
biome check --write

# 【最速】スペルミス・誤字脱字の自動修正 (typos)
typos -w

# 【衛生】不要な export や孤立ファイルの自動削除 (Knip)
npx knip --fix
```

### 個別・検証のみ (CI / 差分確認用)
```bash
# Biome リント検証のみ
biome check

# スペルチェック検証のみ
typos

# Knip デッドコード監査
npx knip
```

---

## 3. ADR-0013 支援ツール群の全体構成

| ツール | 実装 | 設定ファイル | 役割 | 実行コマンド |
|---|---|---|---|---|
| **Biome** | Rust | [`biome.json`](../../../biome.json) | 超高速フォーマッター & リンター | `biome check --write` |
| **typos** | Rust | [`_typos.toml`](../../../_typos.toml) | スペルチェッカー | `typos -w` |
| **ast-grep** (`sg`) | Rust | [`sgconfig.yml`](../../../sgconfig.yml) | 構造的 AST パターン検索・一括置換 | `ast-grep scan` |
| **Repomix** | Node.js | [`repomix.config.json`](../../../repomix.config.json) | AI コンテキスト圧縮・集約 | `repomix` |
| **Knip** | Node.js | [`knip.json`](../../../knip.json) | デッドコード・ゾンビ依存駆除 | `npx knip --fix` |
| **ts-pattern** | TS | (frontend 依存) | 網羅的パターンマッチング（出題状態遷移） | `.exhaustive()` |
| **ts-morph** | TS | (frontend 依存) | TypeScript AST 一括変換・リファクタ | スクリプト実行 |

---

## 4. トラブルシュート・Tips

- **Biome 未インストール環境**: `brew install biome` または `npx @biomejs/biome check --write`。
- **typos 未インストール環境**: `brew install typos-cli`。
- **ast-grep / Repomix**: `brew install ast-grep repomix` で配備可能。
- **固有名詞の誤検知**: メンバー名や独自接頭辞がスペルエラーになった場合は、[`_typos.toml`](../../../_typos.toml) の `[default.extend-words]` に登録する（手動でスペルを崩さない）。

