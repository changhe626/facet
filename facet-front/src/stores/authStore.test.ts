import { describe, it, expect, beforeEach } from 'vitest';
import { useAuthStore } from './authStore';

const testUser = { id: 'u1', email: 'a@b.com', displayName: 'Test', role: 'tenant_admin', createdAt: '2024-01-01' };

describe('authStore', () => {
  beforeEach(() => {
    localStorage.clear();
    useAuthStore.setState({ token: null, user: null, isAuthenticated: false });
  });

  it('starts unauthenticated', () => {
    const state = useAuthStore.getState();
    expect(state.token).toBeNull();
    expect(state.user).toBeNull();
    expect(state.isAuthenticated).toBe(false);
  });

  it('login sets token and user', () => {
    useAuthStore.getState().login('test-token', testUser);
    const state = useAuthStore.getState();
    expect(state.token).toBe('test-token');
    expect(state.user).toEqual(testUser);
    expect(state.isAuthenticated).toBe(true);
    expect(localStorage.getItem('token')).toBe('test-token');
  });

  it('logout clears state', () => {
    useAuthStore.getState().login('t', testUser);
    useAuthStore.getState().logout();
    const state = useAuthStore.getState();
    expect(state.token).toBeNull();
    expect(state.user).toBeNull();
    expect(state.isAuthenticated).toBe(false);
  });

  it('updateUser updates user in state and localStorage', () => {
    const updated = { ...testUser, displayName: 'Updated' };
    useAuthStore.getState().login('t', testUser);
    useAuthStore.getState().updateUser(updated);
    const state = useAuthStore.getState();
    expect(state.user).toEqual(updated);
    expect(JSON.parse(localStorage.getItem('user')!)).toEqual(updated);
  });
});