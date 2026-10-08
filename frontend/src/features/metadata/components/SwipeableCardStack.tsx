'use client';

import { Box } from '@mantine/core';
import {
  animate,
  motion,
  type PanInfo,
  useMotionValue,
  useReducedMotion,
  useTransform,
} from 'motion/react';
import { type ReactNode, useEffect, useRef, useState } from 'react';

interface SwipeableCardStackProps {
  /**
   * 現在表示しているメインのカード要素
   */
  children: ReactNode;
  /**
   * 背後にチラ見せするカード要素（次のカード）
   */
  backgroundCard?: ReactNode;
  /**
   * 次のアイテムが存在するかどうか
   */
  hasNext: boolean;
  /**
   * スワイプ完了（次へ進む）時のコールバック
   */
  onSwipeNext: () => void;
  /**
   * ジェスチャーを無効化するかどうか（モーダル表示中など）
   */
  disabled?: boolean;
}

// スワイプ確定と判定する閾値
const SWIPE_THRESHOLD_X = 85; // 85px 以上水平ドラッグ
const SWIPE_VELOCITY_X = 400; // または 400px/s 以上のフリック速度
const TAP_THRESHOLD_PX = 10; // 10px 未満の移動はタップと判定

export function SwipeableCardStack({
  children,
  backgroundCard,
  hasNext,
  onSwipeNext,
  disabled = false,
}: SwipeableCardStackProps) {
  const shouldReduceMotion = useReducedMotion();
  const x = useMotionValue(0);
  const y = useMotionValue(0);

  const [isPressing, setIsPressing] = useState(false);

  // 指の水平移動量に応じた回転角（-12deg 〜 +12deg）
  const rotate = useTransform(x, [-260, 0, 260], [-12, 0, 12]);

  // 背後カードのスケール拡大（手前のカードが動くと 0.94 -> 1.0 に近づく）
  const bgScale = useTransform(
    x,
    [-200, 0, 200],
    shouldReduceMotion ? [1, 1, 1] : [0.98, 0.94, 0.98],
  );
  const bgOpacity = useTransform(x, [-200, 0, 200], [0.95, 0.6, 0.95]);

  // スワイプ判定の追跡用
  const isDraggingRef = useRef(false);
  const startPosRef = useRef({ x: 0, y: 0 });

  // カード切り替え時に位置をゼロリセット
  useEffect(() => {
    x.set(0);
    y.set(0);
    isDraggingRef.current = false;
    setIsPressing(false);
  }, [x, y]);

  const handlePointerDown = (e: React.PointerEvent) => {
    if (disabled) return;
    startPosRef.current = { x: e.clientX, y: e.clientY };
    isDraggingRef.current = false;
    setIsPressing(true);
  };

  const handlePointerUp = () => {
    setIsPressing(false);
  };

  const handlePanStart = () => {
    isDraggingRef.current = true;
  };

  const handlePan = (_: PointerEvent, info: PanInfo) => {
    if (disabled) return;

    let deltaX = info.offset.x;
    let deltaY = info.offset.y;

    // リスト末尾（次がない）で引っ張る場合は強い抵抗（Rubber-band effect）
    if (!hasNext) {
      deltaX = deltaX * 0.25;
      deltaY = deltaY * 0.25;
    }

    x.set(deltaX);
    y.set(deltaY);
  };

  const handlePanEnd = (_: PointerEvent, info: PanInfo) => {
    setIsPressing(false);
    if (disabled) return;

    const offsetX = info.offset.x;
    const offsetY = info.offset.y;
    const movedDistance = Math.hypot(offsetX, offsetY);
    const velocityMagnitude = Math.hypot(info.velocity.x, info.velocity.y);

    if (movedDistance < TAP_THRESHOLD_PX) {
      // 10px 未満はタップ扱い（位置をリセット）
      animate(x, 0, { type: 'spring', stiffness: 450, damping: 30 });
      animate(y, 0, { type: 'spring', stiffness: 450, damping: 30 });
      isDraggingRef.current = false;
      return;
    }

    // どの方向（左右上下）に払っても、十分な距離またはフリック速度があれば「次のメンバー」へ進む
    const isSwipingNext =
      (movedDistance >= SWIPE_THRESHOLD_X ||
        velocityMagnitude >= SWIPE_VELOCITY_X) &&
      hasNext;

    if (isSwipingNext) {
      // 払った方向へカードを投げる (イグジット)
      const exitDistance = shouldReduceMotion ? 250 : 550;
      const angle = Math.atan2(offsetY, offsetX);
      const targetExitX = Math.cos(angle) * exitDistance;
      const targetExitY = Math.sin(angle) * exitDistance;

      animate(x, targetExitX, {
        type: 'spring',
        stiffness: 300,
        damping: 25,
      });
      animate(y, targetExitY, {
        type: 'spring',
        stiffness: 300,
        damping: 25,
      }).then(() => {
        onSwipeNext();
        x.set(0);
        y.set(0);
        isDraggingRef.current = false;
      });
    } else {
      // 閾値未満または末尾（次がない）の場合はスプリングで中央へ復帰
      animate(x, 0, {
        type: 'spring',
        stiffness: 350,
        damping: 25,
      });
      animate(y, 0, {
        type: 'spring',
        stiffness: 350,
        damping: 25,
      }).then(() => {
        isDraggingRef.current = false;
      });
    }
  };

  return (
    <Box
      style={{
        position: 'relative',
        width: '100%',
        height: '100%',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        overflow: 'visible', // 全面表示: 外側に隠れずカード全体が画面手前に見える
        touchAction: 'none', // 360度自由操作のためブラウザデフォルトタッチを抑止
        userSelect: 'none',
      }}
      onPointerDownCapture={handlePointerDown}
      onPointerUpCapture={handlePointerUp}
      onPointerCancelCapture={handlePointerUp}
    >
      {/* 背後カード (チラ見せ表示、フォーカス・アクセシビリティ対象外) */}
      {backgroundCard && (
        <motion.div
          aria-hidden="true"
          // @ts-expect-error standard HTML inert attribute
          inert=""
          style={{
            position: 'absolute',
            width: '100%',
            height: '100%',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 1,
            scale: bgScale,
            opacity: bgOpacity,
            pointerEvents: 'none',
          }}
        >
          {backgroundCard}
        </motion.div>
      )}

      {/* 手前カード (360度自由ドラッグ・押し込み光沢・浮遊感) */}
      <motion.div
        drag={!disabled} // 360度どこでも運べる全方向ドラッグ
        dragConstraints={{ left: 0, right: 0, top: 0, bottom: 0 }}
        dragElastic={1} // 制約なしの完全自由追従
        onPanStart={handlePanStart}
        onPan={handlePan}
        onPanEnd={handlePanEnd}
        animate={{
          scale: isPressing && !disabled ? 1.02 : 1,
          boxShadow:
            isPressing && !disabled
              ? '0 24px 48px -12px rgba(0, 0, 0, 0.5), 0 0 20px rgba(255, 255, 255, 0.15)'
              : '0 8px 24px -8px rgba(0, 0, 0, 0.3)',
        }}
        transition={{ type: 'spring', stiffness: 400, damping: 25 }}
        style={{
          position: 'relative',
          width: '100%',
          height: '100%',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          zIndex: 10, // 全面表示
          x,
          y,
          rotate: shouldReduceMotion ? 0 : rotate,
          cursor: disabled ? 'default' : 'grab',
          borderRadius: 16,
        }}
        whileTap={{ cursor: disabled ? 'default' : 'grabbing' }}
      >
        {children}

        {/* 押し込み・掴み時の光沢感 (シマーハイライト・グロスオーバーレイ) */}
        <motion.div
          animate={{ opacity: isPressing && !disabled ? 1 : 0 }}
          transition={{ duration: 0.15 }}
          style={{
            position: 'absolute',
            inset: 0,
            borderRadius: 16,
            background:
              'linear-gradient(135deg, rgba(255,255,255,0.2) 0%, rgba(255,255,255,0.03) 40%, rgba(0,0,0,0.1) 100%)',
            boxShadow: 'inset 0 0 0 1.5px rgba(255, 255, 255, 0.35)',
            pointerEvents: 'none',
            zIndex: 30,
          }}
        />
      </motion.div>
    </Box>
  );
}
