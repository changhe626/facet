import { create } from 'zustand';
import { getActiveEnvironmentsApi } from '@/api/environment';
import type { Environment } from '@/types';

interface EnvState {
  environments: Environment[];
  currentEnvId: string | null;
  loading: boolean;
  fetchEnvironments: () => Promise<void>;
  setCurrentEnv: (envId: string) => void;
}

export const useEnvStore = create<EnvState>((set) => ({
  environments: [],
  currentEnvId: null,
  loading: false,

  fetchEnvironments: async () => {
    set({ loading: true });
    try {
      const res = await getActiveEnvironmentsApi();
      const envs: Environment[] = res.data || [];
      set((state) => ({
        environments: envs,
        currentEnvId: state.currentEnvId || (envs.length > 0 ? envs[0].id : null),
        loading: false,
      }));
    } catch {
      set({ loading: false });
    }
  },

  setCurrentEnv: (envId: string) => set({ currentEnvId: envId }),
}));