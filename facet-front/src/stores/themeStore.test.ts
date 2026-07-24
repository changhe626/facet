import { describe, it, expect, beforeEach } from 'vitest';
import { useThemeStore } from './themeStore';

describe('themeStore', () => {
  beforeEach(() => {
    localStorage.clear();
    useThemeStore.setState({ mode: 'light' });
  });

  it('defaults to light mode', () => {
    expect(useThemeStore.getState().mode).toBe('light');
  });

  it('toggleTheme switches between light and dark', () => {
    useThemeStore.getState().toggleTheme();
    expect(useThemeStore.getState().mode).toBe('dark');
    expect(localStorage.getItem('theme')).toBe('dark');

    useThemeStore.getState().toggleTheme();
    expect(useThemeStore.getState().mode).toBe('light');
    expect(localStorage.getItem('theme')).toBe('light');
  });

  it('setTheme sets the mode', () => {
    useThemeStore.getState().setTheme('dark');
    expect(useThemeStore.getState().mode).toBe('dark');
    expect(localStorage.getItem('theme')).toBe('dark');
  });
});