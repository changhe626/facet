import { create } from 'zustand';
import { STORAGE_KEYS } from '@/utils/constants';
import type { User } from '@/types';

interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  login: (token: string, user: User) => void;
  logout: () => void;
  updateUser: (user: User) => void;
}

const getUser = (): User | null => {
  try {
    const u = localStorage.getItem(STORAGE_KEYS.USER);
    return u ? JSON.parse(u) : null;
  } catch {
    return null;
  }
};

const getToken = (): string | null => localStorage.getItem(STORAGE_KEYS.TOKEN);

export const useAuthStore = create<AuthState>((set) => ({
  token: getToken(),
  user: getUser(),
  isAuthenticated: !!getToken(),

  login: (token: string, user: User) => {
    localStorage.setItem(STORAGE_KEYS.TOKEN, token);
    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user));
    set({ token, user, isAuthenticated: true });
  },

  logout: () => {
    localStorage.removeItem(STORAGE_KEYS.TOKEN);
    localStorage.removeItem(STORAGE_KEYS.USER);
    set({ token: null, user: null, isAuthenticated: false });
  },

  updateUser: (user: User) => {
    localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user));
    set({ user });
  },
}));