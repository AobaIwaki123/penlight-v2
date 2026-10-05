---
id: ADR-0028
title: ADR コンプライアンステスト駆動によるトレーサビリティおよび二重管理防止アーキテクチャ
status: Accepted
scope: System
primary_category: DEV
categories: [DEV, APP]
tags: [adr-governance, toolchain, yagni]
deciders: [user, ai]
date: 2026-10-05
---

# 0028. ADR コンプライアンステスト駆動によるトレーサビリティおよび二重管理防止アーキテクチャ (0028-compliance-test-driven-adr-traceability-architecture.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-05

---

## 1. 背景と解決すべき課題 (Context & Problem)

本プロジェクトではアーキテクチャの意思決定を ADR（`adr/`）で管理し、累計 20 件以上の設計決定が蓄積されています。
しかし、開発が進むにつれて以下の課題が顕在化しました：

1. **実装状況の不透明性**: ADR 一覧（`adr/README.md`）には設計ステータス（Accepted / Proposed）しかなく、コード側でどこまで実装されているのかがコードを読まないと判別できない。
2. **手動ドキュメント管理による二重管理の破綻**: Markdown に「実装済み / 未実装」を手動記述すると、コードが修正された瞬間にドキュメントが陳腐化して嘘をつくようになる。
3. **過剰な自動化ツールの保守コスト (YAGNI 違反)**: ソースコードの AST 解析やカスタムリントを独自構築してドキュメントを自動生成する仕組みは、開発および保守のコストが過大となる。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

テストコードを **「唯一の動く仕様書（Executable Specification as Single Source of Truth）」** と位置づけ、テストの実行結果をもって ADR の実装・仕様準拠を機械的に証明する **ADR コンプライアンステスト駆動アーキテクチャ** を採用する。

```mermaid
flowchart LR
    ADR["ADR 意思決定<br/>(adr/00XX.md)"] --> Test["ADR Compliance Test<br/>(t.Run('ADR-00XX: ...'))"]
    Test --> Pass["✅ Test PASS<br/>(= 実装完了・仕様準拠の証明)"]
    Test --> Fail["❌ Test FAIL / 存在せず<br/>(= 未実装またはリグレッション)"]
```

### 具現化仕様

1. **ADR 具現化仕様のナンバリング規約 (Specification Numbering)**:
   - 今後作成・改訂される ADR の「2. 決定事項と具現化仕様」は、必ず連番（`1.`, `2.`, `3.`）で項目化し、先頭に対象レイヤー（`[Backend]`, `[Frontend]`, `[DB]`, `[Contract]` 等）を明記する。
   - 例:
     ```markdown
     ## 2. 決定事項と具現化仕様
     1. [Backend] 母集団フィルタリング関数 (FilterMembers) の提供
     2. [Frontend] FilterModal における衣装 (PhotoType) 選択 UI の提供
     3. [Contract] 候補者不足時の Problem Details (ERR_INVALID_INPUT) 返却
     ```

2. **サブ要件テスト命名規約 (`ADR-00XX/Y`)**:
   - 各テストケース名には、対応する ADR 番号とサブ要件番号（`ADR-00XX/Y`）を冠する。
   - バックエンド (Go):
     ```go
     t.Run("ADR-0018/1: [Backend] Candidate Pool Filtering Logic", func(t *testing.T) { ... })
     t.Run("ADR-0018/3: [Contract] Error on Insufficient Candidates", func(t *testing.T) { ... })
     ```
   - フロントエンド (TypeScript / Vitest 等):
     ```ts
     test('ADR-0018/2: [Frontend] PhotoType Selection UI in FilterModal', () => { ... });
     ```

3. **「一部未実装 (Partial)」の客観的判定ルール**:
   - ADR 内の全サブ要件（1〜N）のテストが存在し、すべて PASS ──► **✅ 完全実装 (Implemented)**
   - 一部のサブ要件のみテストが存在・PASS（例: Backend のみ PASS、Frontend が未実装または `test.todo` / `t.Skip`） ──► **⚠️ 一部実装 (Partial)**
   - テストが 1 つも存在しない、またはテストが失敗 ──► **❌ 未実装 (Unimplemented)**

4. **移行・整理プロセスの分離**:
   - 既存の全 ADR に対する現時点の静的な実装達成率マトリクスは、一旦 `docs/notes/08_adr_implementation_status_and_achievement_matrix.md` で一時的に可視化・整理する。
   - 各 ADR の機能実装時に、本規約に基づくコンプライアンステストを追加していくことで、自動的に静的ノートからテスト駆動の動的保証へ移行させる。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### 実装追跡手法の比較検討

| 方式 | 二重管理リスク | 導入・保守コスト | 信頼性・精度 | 判定 |
|---|---|---|---|---|
| **A. 静的 Markdown 手動管理** | **極めて高い**（即時陳腐化） | 低い | **低い**（人間の更新忘れで嘘になる） | **却下** |
| **B. AST 解析・独自リント生成** | ゼロ | **極めて高い**（独自ツールの保守地獄） | 高い | **却下**: YAGNI 違反 |
| **C. ADR Compliance Test [採択]** | **ゼロ**（コードが唯一の正本） | **最小限**（標準テストランナー利用） | **最高**（CI で常時自動検証） | **採用** |

- **決定打**: テストが通っていること自体が「設計通りに動き続けている」最高の証跡である。ドキュメントを更新する作業を完全撤廃でき、AI エージェントにとっても「ADR 番号を冠したテストをパスさせる」という明確なゴール（TDD）が成立する。

---

## 4. 得られる効果と今後の展望 (Consequences)

### ポジティブな影響 (Positive)
- **完全なメンテフリー**: 実装状況を更新するためのドキュメント修正作業がゼロになる。
- **リグレッションの機械的防止**: 後からの変更で過去の ADR 仕様が破壊された場合、テスト失敗として即座に検知される。
- **ワンコマンドでの状況把握**: `go test -v ./... | grep "ADR-"` 等により、現在動作保証されている ADR をターミナルから 1 秒でリストアップ可能。

### 留意点と対策 (Negative & Mitigation)
- **網羅率の段階的引き上げ**:
  - 過去の ADR すべてに初日からテストを書くのは工数がかかるため、未実装機能の実装時および既存改修時に順次コンプライアンステストを拡充する。
