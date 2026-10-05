---
id: ADR-0027
title: 楽曲ペンライトカラー（1色/2色・左右なし）データ構造の採用
status: Accepted
scope: Domain
primary_category: DOM
categories: [DOM, APP]
tags: [single-source-of-truth, typeid, quiz-strategy]
deciders: [user, ai]
date: 2026-10-05
---

# 0027. 楽曲ペンライトカラー（1色/2色・左右なし）データ構造の採用 (0027-song-penlight-color-data-structure.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-05

---

## 1. 背景と解決すべき課題 (Context & Problem)

本システムでは、メンバー個人のペンライトカラー（左右2本）を中心に出題・管理を行ってきた。
しかし、ライブ会場では特定の楽曲において観客全体で統一したペンライトカラーを点灯させる文化（例: 日向坂46「月と星が踊るMidnight」= 青×青、櫻坂46「Start over!」= 白×赤、=LOVE「絶対アイドル辞めないで」等）が存在する。

これらをクイズとして出題・管理するため、新たなエンティティが必要となった。
その際、客席ブロック別の虹色分けやセンター推しメン連動といった過剰に複雑な例外仕様を持ち込まず、保守性と出題アルゴリズムのシンプルさを維持したデータ構造の策定が求められる。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

以下の割り切りと原則に基づく `Song` エンティティを正式採用する。

### 1. `Song` エンティティの定義 (TypeID: `sng_<uuidv7>`)

```go
// Song represents a musical track entity and its official/live penlight colors.
type Song struct {
	ID        ID        `json:"id"`          // sng_... (UUID v7)
	GroupID   ID        `json:"group_id"`    // 所属グループ (grp_...)
	Title     string    `json:"title"`       // 楽曲タイトル (例: "絶対アイドル辞めないで")
	Kana      *string   `json:"kana"`        // 読み仮名 (任意・未設定可, ソート補助用)
	Color1ID  ID        `json:"color1_id"`   // 1色目 (必須, col_...)
	Color2ID  *ID       `json:"color2_id"`   // 2色目 (任意, col_...。1色の曲は nil)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

### 2. 3 つの設計原則
1. **左右の概念を排除 (No Left/Right Concept)**:
   - メンバー個人のカラーは「左手・右手」の持ち手が存在するが、楽曲カラーは「会場全体を構成する色」であり、左右の区別は存在しない。
   - クイズ判定においても、解答された 2 色は順序を問わない無順序集合として正誤判定を行う。
2. **1 色または 2 色への限定（デフォルト 1 色）**:
   - `Color1ID` を必須（デフォルト 1 色目）とし、2 色指定の楽曲のみ `Color2ID` に値を保持する。
   - `Color2ID == nil` の場合は 1 色指定（全席同色）として扱う。
3. **例外演出の非サポート (YAGNI 原則)**:
   - 会場ブロック別の虹色分けや、センターメンバーの推しメンカラーに自動追従するような動的・例外的仕様は一切サポートしない。
   - 演出メモ用の `description` カラムも不要とし、純粋な識別・カラー関連のみに絞る。
   - `Kana` は英字・記号混じりの曲名における入力コストを考慮し、任意（Optional: `*string`）とする。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### データ構造アプローチの比較検討

| 評価軸 | 採用案（1色/2色・左右なし） | 代替案（左右固定・多色対応・例外サポート） |
|---|---|---|
| **スキーマの簡潔さ** | **極めて高い**（外部キー 2 個のみ） | **低い**（中間テーブル、ColorType Enum等が必要） |
| **判定ロジック** | **シンプル**（1色一致 or 2色無順序一致） | **複雑**（ブロック判定、センター連動クエリ等） |
| **運用・入力負荷** | **最小**（タイトルと1〜2色を選ぶだけ） | **大**（演出種別や注記の管理が必要） |
| **表現力** | **95% 以上の通常楽曲を網羅** | 特殊演出（虹色等）も表現可能だが運用負担大 |

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響
- **クイズ出題形式の拡張 ([ADR-0019](./0019-quiz-format-strategy-and-color-palette-architecture.md))**:
  - 「楽曲名 → ペンライトカラー」当てクイズ、または「カラー → 楽曲名」当てクイズを新規 Strategy として容易にプラグイン可能。
- **データ不整合の防止**:
  - カラー ID（`col_...`）への外部キー参照により、カラー名変更時も楽曲データが自動追従する（[ADR-0006](./0006-domain-schema-and-typeid-structure.md) 準拠）。

### 留意点と対策
- **1色曲と2色曲の出題UI**:
  - クイズ出題時、「1 色を選択させる問題」と「2 色を選択させる問題」で解答 UI の入力を適切に切り替える（または、常に 2 本点灯として同色選択を許容する）。
