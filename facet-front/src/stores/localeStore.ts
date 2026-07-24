import { create } from 'zustand';
import { en } from '@/locales/en';
import { zh } from '@/locales/zh';

export type Locale = 'en' | 'zh';
type TranslationMap = typeof en;

const translations: Record<Locale, TranslationMap> = { en, zh };

interface LocaleState {
  locale: Locale;
  t: TranslationMap;
  setLocale: (locale: Locale) => void;
}

  const stored = localStorage.getItem('locale') as Locale | null;
  const initial = stored === 'zh' ? 'zh' : 'en';

export const useLocaleStore = create<LocaleState>((set) => ({
  locale: initial,
  t: translations[initial],
  setLocale: (locale: Locale) => {
    localStorage.setItem('locale', locale);
    set({ locale, t: translations[locale] });
  },
}));
