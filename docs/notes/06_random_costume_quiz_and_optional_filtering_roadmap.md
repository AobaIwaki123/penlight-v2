# 06. 衣装別ランダム出題およびオプショナル衣装フィルタリング開発ロードマップ

- **ステータス**: 検討中 (Draft Roadmap)
- **日付**: 2026-10-05
- **関連 ADR**:
  - [ADR-0018: クイズ出題における母集団フィルタリング設計の分離](../../adr/0018-quiz-candidate-pool-filtering-architecture.md)
  - [ADR-0020: 出題対象メンバー選出におけるブレンドデッキ戦略の採用](../../adr/0020-quiz-target-member-selection-strategy.md)
  - [ADR-0021: GitOps 宣言的マスタデータ同期・バージョン管理アーキテクチャ](../../adr/0021-gitops-master-data-synchronization-and-versioning-architecture.md)
  - [ADR-0022: プラガブル出題 UI・プレゼンテーションレイアウト設計](../../adr/0022-pluggable-quiz-ui-and-presentation-layout-architecture.md)

---

## 1. 背景と現状の課題

バックエンド（`pkg/quiz`）では、ADR-0018 / ADR-0020 に基づき「衣装指定時は該当衣装から、指定なし時は全衣装からランダムに画像を選出する」ロジック（`quiz.SelectQuestionImage`）およびテストが実装済みである。
しかし、Local-First（ADR-0007）で動作するフロントエンド（`frontend/src/features/quiz/`）では、以下の課題が存在する：

1. **出題画像が先頭 1 枚に固定**:
   - `QuizContainer.tsx` および各レイアウトコンポーネント（Classic, Compact, Overlay）が `target.images?.[0]` を直接参照しており、メンバーが複数の衣装画像（`PhotoType`）を持っていても常に初期画像しか表示されない。
2. **衣装絞り込み UI の未提供**:
   - `FilterModal.tsx` にはグループ・期生・卒業生の選択肢しか存在せず、特定衣装（例: 「13th Single 制服」など）に絞り込んでプレイする手段がない。

---

## 2. 要件仕様

1. **デフォルト動作（衣装指定なし）**:
   - 出題ごとに、該当メンバーが保有する画像（`member.images`）の中からランダムに 1 枚を選定して表示する。
   - 表示される衣装名（`costumeTitle`）も、選定された画像の `photo_type.name` に動的連動させる。
2. **衣装フィルタリング（Optional）**:
   - `FilterModal` に「衣装指定（任意）」の選択肢を追加する。
   - デフォルトは「指定なし（全衣装からランダム）」。
   - 特定の衣装が選択された場合は、その衣装を持つメンバーのみを母集団とし、出題画像も該当衣装から選出する。

---

## 3. 技術設計

### 3.1. 出題アイテムとレイアウト Props の拡張
出題中の画像再抽選チラつきを防止するため、デッキ生成（または問題確定）時にメンバーと選出画像をペアにしたデータ構造を保持する。

```ts
// frontend/src/features/quiz/types.ts
export interface QuizDeckItem {
  member: Member;
  selectedImage?: MemberImage;
}

export interface TargetLayoutProps {
  target: Member;
  selectedImage?: MemberImage; // 追加
  costumeTitle: string;
  selectedLeftColor?: Color;
  selectedRightColor?: Color;
  isCorrect?: boolean;
  onOpenInput?: (hand: 'left' | 'right') => void;
}
```

### 3.2. ブラウザ内画像選出関数
Go 側の `quiz.SelectQuestionImage`（`pkg/quiz/deck.go`）と完全に一致するロジックを TypeScript 側に関数化する。

```ts
export function selectQuestionImage(
  member: Member,
  photoTypeIds?: string[],
): MemberImage | undefined {
  if (!member.images || member.images.length === 0) return undefined;

  if (photoTypeIds && photoTypeIds.length > 0) {
    const matched = member.images.filter((img) =>
      photoTypeIds.includes(img.photo_type_id)
    );
    if (matched.length > 0) {
      return matched[Math.floor(Math.random() * matched.length)];
    }
  }

  return member.images[Math.floor(Math.random() * member.images.length)];
}
```

### 3.3. レイアウトコンポーネントの表示切り替え
`LayoutClassic`, `LayoutCompact`, `LayoutOverlay` の各画像参照部を以下に統一：

```tsx
const displayImg = selectedImage || target.images?.[0];
const imageSrc = getImageUrl(displayImg?.image_key) || fallbackUrl;
```

### 3.4. FilterModal への衣装選択 UI 追加
- 選択中グループのメンバーが保持する `photo_type` のユニーク一覧を抽出し、ソートして選択肢を生成。
- 「指定なし（全衣装）」を初期選択値とし、オプショナルに特定衣装を選択可能にする。

---

## 4. 実装ステップと検証計画

1. **ステップ 1: 型定義 & 出題ロジックの実装**
   - `QuizDeckItem` 定義および `selectQuestionImage` 関数の作成。
   - `QuizContainer.tsx` のデッキ初期化・リスタート処理での画像ランダム割り当て。
2. **ステップ 2: レイアウトコンポーネントへの適用**
   - Overlay, Compact, Classic の各コンポーネントで `selectedImage` を表示に反映。
3. **ステップ 3: FilterModal への衣装選択 UI 追加**
   - グループ別 PhotoType 抽出とフィルタ適用処理の追加。
4. **ステップ 4: 検証**
   - `make verify-ai`（Biome, Knip, 型チェック）のパス確認。
   - ブラウザ上で出題ごとに衣装が切り替わること、特定衣装フィルタ時に正しく絞り込まれることを確認。
