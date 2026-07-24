import { create } from 'zustand';
import { STORAGE_KEYS } from '@/utils/constants';

type ThemeMode = 'light' | 'dark';

interface ThemeState {
  mode: ThemeMode;
  toggleTheme: () => void;
  setTheme: (mode: ThemeMode) => void;
}

const getInitial = (): ThemeMode => {
  const stored = localStorage.getItem(STORAGE_KEYS.THEME);
  if (stored === 'dark' || stored === 'light') return stored;
  return 'light';
};

export const useThemeStore = create<ThemeState>((set) => ({
  mode: getInitial(),

  toggleTheme: () =>
    set((state) => {
      const next: ThemeMode = state.mode === 'light' ? 'dark' : 'light';
      localStorage.setItem(STORAGE_KEYS.THEME, next);
      return { mode: next };
    }),

  setTheme: (mode: ThemeMode) => {
    localStorage.setItem(STORAGE_KEYS.THEME, mode);
    set({ mode });
  },
}));