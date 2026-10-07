/**
 * Lightweight feature switches.
 * Easily toggle experimental or optional feature provisions ON/OFF.
 */
export const FEATURES = {
  /**
   * オフラインモードでの画像全件フェッチ機能提供フラグ
   * true: ヘッダー/メニューにオフライン準備ボタンを表示
   * false: 手動フェッチUIを非表示（通常プレイでの表示済み画像キャッシュは継続動作）
   */
  ENABLE_OFFLINE_IMAGE_FETCH: true,
} as const;
