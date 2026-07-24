import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useEnvStore } from './envStore';

const mockGetActive = vi.fn();
vi.mock('@/api/environment', () => ({
  getActiveEnvironmentsApi: () => mockGetActive(),
}));

describe('envStore', () => {
  beforeEach(() => {
    useEnvStore.setState({ environments: [], currentEnvId: null, loading: false });
    mockGetActive.mockReset();
  });

  it('starts with empty state', () => {
    const state = useEnvStore.getState();
    expect(state.environments).toEqual([]);
    expect(state.currentEnvId).toBeNull();
    expect(state.loading).toBe(false);
  });

  it('fetchEnvironments sets environments and currentEnvId', async () => {
    mockGetActive.mockResolvedValue({ data: [{ id: 'env-1', name: 'Prod' }, { id: 'env-2', name: 'Staging' }] });
    await useEnvStore.getState().fetchEnvironments();
    const state = useEnvStore.getState();
    expect(state.environments).toHaveLength(2);
    expect(state.currentEnvId).toBe('env-1');
    expect(state.loading).toBe(false);
  });

  it('setCurrentEnv updates selected env', () => {
    useEnvStore.getState().setCurrentEnv('env-2');
    expect(useEnvStore.getState().currentEnvId).toBe('env-2');
  });
});