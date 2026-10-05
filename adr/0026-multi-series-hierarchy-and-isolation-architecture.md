---
id: ADR-0026
title: 同一アプリ内におけるシリーズ（Series）階層分離アーキテクチャの採用
status: Accepted
scope: System
primary_category: ARC
categories: [ARC, DOM]
tags: [extensible-group, surrogate-key, directory-structure]
deciders: [user, ai]
date: 2026-10-05
---

# 0026. 同一アプリ内におけるシリーズ（Series）階層分離アーキテクチャの採用 (0026-multi-series-hierarchy-and-isolation-architecture.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-05

---

## 1. 背景と解決すべき課題 (Context & Problem)

本システムはこれまで坂道グループ（乃木坂46・櫻坂46・日向坂46）を対象に運用されてきた。
今回、新たに =LOVE 系列（=LOVE、≠ME、≒JOY 等）を追加導入する要件が生じた。

アイドルファン文化において、異なる系列（坂道シリーズ vs =LOVE 系列）はファン層や文脈が大きく異なり、**「クイズ出題やカラーパレット、グループ選択において決して混ざり合わないこと」** が必須要件となる。

一方、これを別アプリ（別 Pod / 別 DB / 別ドメイン）として物理分離した場合、自宅 Kubernetes クラスタにおける Pod・PVC・Service・Ingress、および CI/CD パイプラインが 2 重化し、単一バイナリ・軽量運用（メモリ 32MiB）という基本設計方針（[ADR-0002](./0002-backend-go-architecture.md), [ADR-0012](./0012-kubernetes-deployment-and-gitops-architecture.md)）を損なう。

---

## 2. 決定事項と具現化仕様 (Decision & Specification)

単一バイナリ・単一 DB 構成を維持したまま、**同一アプリ内における「シリーズ (Series)」階層分離** を正式採用する。

### 1. `Series` エンティティの新設 (TypeID: `ser_<uuidv7>`)
グループ（`Group`）の上位概念として `Series` を定義する。

```go
// Series represents an idol franchise or series (e.g. "坂道シリーズ", "=LOVE系列").
type Series struct {
	ID           ID        `json:"id"`            // ser_... (UUID v7)
	Name         string    `json:"name"`          // Formal name, e.g. "坂道シリーズ", "=LOVE系列"
	Slug         string    `json:"slug"`          // URL-safe identifier, e.g. "sakamichi", "ikolove"
	DisplayOrder int       `json:"display_order"` // UI sort order
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
```

### 2. グループとの関連付け
- `Group` エンティティに `SeriesID ID` (`ser_...`) を追加。
- 乃木坂46・櫻坂46・日向坂46は `ser_sakamichi`、=LOVE・≠ME・≒JOY は `ser_ikolove` に所属する。

### 3. 出題エンジンおよびパレットのシリーズ境界強制 ([ADR-0018](./0018-quiz-candidate-pool-filtering-architecture.md))
- 出題母集団フィルタ（`pkg/quiz/filter.go`）の最上位条件として `SeriesID` を強制。
- クイズ生成時に、アクティブなシリーズ以外のグループ・メンバー・公式カラーが候補プールや解答選択肢に混入することを構造的に遮断する。

---

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

### 分離方式の比較検討

| 評価軸 | 選択肢 A: シリーズ階層分離（採用） | 選択肢 B: 別アプリ・別Pod物理分離 |
|---|---|---|
| **分離性（混ざらない保証）** | **高**（出題エンジンのシリーズ強制フィルタで担保） | **最高**（物理DB分離） |
| **インフラ・運用コスト** | **極小**（単一 Pod、メモリ 32MiB 据え置き） | **大**（Pod/PVC/Ingress が 2 重化） |
| **機能改善・更新コスト** | **最小**（1 回のデプロイで両系列に自動反映） | **大**（個別ビルド・個別デプロイ） |
| **拡張性** | **高**（新系列の追加も DB マスタ登録で完了） | **低**（新アプリ・新マニフェスト作成が必要） |

---

## 4. 得られる効果と留意点 (Consequences)

### ポジティブな影響
- **単一バイナリ・軽量運用の維持**: リソース消費やマニフェストを増やすことなく、複数系列のアイドルクイズを単一インスタンスで提供できる。
- **無停止・ノーコード拡張**: 将来 48グループ等の別系列を追加する場合も、コード修正・再デプロイを行わず DB マスタの投入のみで対応可能（[ADR-0006](./0006-domain-schema-and-typeid-structure.md) 準拠）。

### 留意点と対策
- **UI コンテキストの維持**: フロントエンドで現在選択されているシリーズを状態管理（URL パスまたは LocalStorage）し、ユーザーが意図したシリーズのクイズのみが確実に開始されるようルーティングを設計する。
