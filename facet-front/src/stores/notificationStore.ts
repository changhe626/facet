import { create } from 'zustand';
import { getNotificationStatusApi } from '@/api/notifications';

interface NotificationState {
  unreadCount: number;
  fetchUnreadCount: () => Promise<void>;
  setUnreadCount: (count: number) => void;
}

export const useNotificationStore = create<NotificationState>((set) => ({
  unreadCount: 0,

  fetchUnreadCount: async () => {
    try {
      const res = await getNotificationStatusApi();
      set({ unreadCount: res.data?.unreadCount ?? 0 });
    } catch {
    }
  },

  setUnreadCount: (count: number) => set({ unreadCount: count }),
}));