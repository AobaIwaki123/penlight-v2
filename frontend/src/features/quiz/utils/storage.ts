import type { InputMode, LayoutMode } from '@/features/quiz/types';

interface SavedQuizSettings {
  groupId: string;
  songMode: boolean;
}

const STORAGE_KEYS = {
  SETTINGS: 'penlight_quiz_settings_v3',
  LAYOUT: 'penlight_layout_mode',
  INPUT: 'penlight_input_mode',
} as const;

export function loadSavedSettings(): SavedQuizSettings | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = localStorage.getItem(STORAGE_KEYS.SETTINGS);
    if (!raw) return null;
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function saveSettings(settings: SavedQuizSettings): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEYS.SETTINGS, JSON.stringify(settings));
  } catch {
    // Ignore storage quota errors
  }
}

export function loadSavedLayoutMode(): LayoutMode | null {
  if (typeof window === 'undefined') return null;
  try {
    const val = localStorage.getItem(STORAGE_KEYS.LAYOUT);
    if (val === 'classic' || val === 'overlay' || val === 'compact') {
      return val;
    }
  } catch {
    // Ignore
  }
  return null;
}

export function saveLayoutMode(mode: LayoutMode): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEYS.LAYOUT, mode);
  } catch {
    // Ignore
  }
}

export function loadSavedInputMode(): InputMode | null {
  if (typeof window === 'undefined') return null;
  try {
    const val = localStorage.getItem(STORAGE_KEYS.INPUT);
    if (
      val === 'grid' ||
      val === 'donut' ||
      val === 'dots' ||
      val === 'sheet'
    ) {
      return val;
    }
  } catch {
    // Ignore
  }
  return null;
}

export function saveInputMode(mode: InputMode): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEYS.INPUT, mode);
  } catch {
    // Ignore
  }
}
