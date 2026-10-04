import type { Color, Member } from '@/types/generated';

// TargetLayoutProps is the unified interface for presenting member information.
export interface TargetLayoutProps {
  target: Member;
  costumeTitle: string;
  selectedLeftColor?: Color;
  selectedRightColor?: Color;
  isCorrect?: boolean;
}

// AnswerInputProps is the pluggable interface for answer input components.
export interface AnswerInputProps {
  target: Member;
  colors: Color[];
  onAnswer: (input: { leftColorId: string; rightColorId: string }) => void;
  disabled: boolean;
}

// LayoutMode represents the visual presentation style.
export type LayoutMode = 'classic' | 'overlay' | 'compact';
