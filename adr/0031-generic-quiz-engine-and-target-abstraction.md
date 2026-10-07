---
id: ADR-0031
title: クイズ出題・判定エンジンのジェネリック抽象化アーキテクチャの採用
status: Accepted
scope: Backend
primary_category: ARC
categories: [ARC, APP]
tags: [go, quiz-strategy, yagni]
deciders: [user, ai]
date: 2026-10-05
---

# 0031. クイズ出題・判定エンジンのジェネリック抽象化アーキテクチャの採用 (0031-generic-quiz-engine-and-target-abstraction.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-05

---

## 1. 背景と解決すべき課題 (Context & Problem)

本システムにおいて、実際のユーザー出題・判定は端末内の Local-First（TypeScript / `src/features/quiz/logic.ts`）で実行される（Ref: [ADR-0007](./0007-local-first-offline-pwa-architecture.md)）。一方、サーバーサイド（`pkg/quiz/`）はドメイン仕様・コンプライアンステストの正本アルゴリズムおよび将来のサーバー検証・バッチ処理を担っている。

従来の Go 側クイズエンジン（`pkg/quiz/`）は、メンバー推しメンカラークイズに特化して構築されており、以下の強い具象結合が存在していた。

1. **出題対象の具象結合**:
   - `FilterMembers(members []model.Member, ...)`、`BuildBlendedDeck(pool []model.Member, ...)` のように、引数および戻り値がすべて `model.Member` 構造体にハードコードされていた。
2. **正誤判定の左右2色固定**:
   - `JudgeAnswer(color1ID, color2ID, target model.Member)` はメンバーの左右 2 色比較のみを前提としており、[ADR-0027](./0027-song-penlight-color-data-structure.md) で定義された楽曲の「1色指定（全席同色）」に対応できなかった。

楽曲クイズ（`Song`）を導入するにあたり、楽曲専用の別関数（`BuildSongDeck`、`JudgeSong`）をコピペ・量産すると、デッキ構築アルゴリズム（未出題/復習比率、Fisher-Yates シャッフル）の二重管理と技術的負債を引き起こす。
そのため、出題および判定の本質ロジックを抽出・共通化するアーキテクチャが求められる。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

Go のジェネリクス（Generics）および最小限のインターフェースを用いた**出題・判定ロジックの抽象化**を正式採用する。

### 1. `QuizTarget` インターフェースの定義
出題対象となり得るすべてのエンティティ（メンバー、楽曲等）が満たす共通インターフェースを定義する。

```go
// QuizTarget represents an entity that can be quizzed for official penlight colors.
type QuizTarget interface {
	GetID() ID
	GetCorrectColors() []ID // Returns 1 or 2 official color IDs
}
```

- `Member` の実装: `LeftColorID` と `RightColorID` のスライス（2色）を返す。
- `Song` の実装: `Color1ID`（1色）または `Color1ID` と `Color2ID`（2色）のスライスを返す。

### 2. ジェネリクスによる型安全デッキ構築 (`pkg/quiz/deck.go`)
出題デッキ構築関数をジェネリック化し、型安全性を維持しながら任意の `QuizTarget`（`Member` または `Song`）を処理可能とする。

```go
// BuildBlendedDeck creates a quiz deck blending unseen and review items for ANY QuizTarget.
func BuildBlendedDeck[T QuizTarget](
	pool []T,
	history []model.AnswerLog,
	deckSize int,
	rng *rand.Rand,
) []T {
	// 未出題・復習判定、Fisher-Yates シャッフル、重複排除アルゴリズムを共通実行
}
```

### 3. 無順序カラー集合一致（Set Equality）による正誤判定の統一 (`pkg/quiz/judge.go`)
判定ロジックを「左右の持ち手」ではなく、**「正解カラー集合と回答カラー集合の無順序一致判定」** として一般化する。

```go
// JudgeAnswer checks whether user's selected colors match target's correct colors as an unordered set.
func JudgeAnswer(selectedColors []model.ID, target QuizTarget) bool {
	// 要素数の一致、およびスライス間の相互包含（順序不問）を検証
}
```
- 楽曲 1 色問題: 解答 1 色と正解 1 色が一致すれば正解。
- メンバー / 楽曲 2 色問題: 解答された 2 色が順序を問わず正解 2 色と一致すれば正解。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### 設計アプローチの比較検討

| 評価軸 | 採用案（ジェネリック抽象化） | 代替案 A（個別専用エンジンの量産） | 代替案 B（interface{} / Any型） |
|---|---|---|---|
| **コードの保守性** | **極めて高い**（コアロジック単一） | **低い**（コピペによる二重管理） | **中**（キャスト多発でバグ誘発） |
| **コンパイル時型安全性** | **完全**（Go Generics で保証） | **完全** | **皆無**（実行時パニックの危険） |
| **拡張性** | **極めて高い**（新クイズも Target 実装のみ） | **低い**（新規関数追加が必要） | **中** |
| **YAGNI 原則整合性** | **合致**（必要最小限のインターフェース）| **違反**（無駄なコード増殖） | **違反** |

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響
- **DRY 原則の徹底**: 複雑なデッキ構築アルゴリズム（未出題比率 7:3、Fisher-Yates）の単一正本性が保たれる。
- **高テスト容易性**: `QuizTarget` のシンプルなモック構造体を作成するだけで、出題エンジン全パターンの単体テストを網羅できる。

### 留意点と対策
- **回答ログのターゲット識別**: 履歴テーブル（`AnswerLog`）において、回答対象がメンバーか楽曲かを識別できるよう、対象サロゲートキー（TypeID: `mem_...` vs `sng_...`）のプレフィックス判定を活用して履歴照合を行う。
