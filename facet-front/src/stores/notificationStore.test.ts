import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useNotificationStore } from './notificationStore';

const mockGetStatus = vi.fn();
vi.mock('@/api/notifications', () => ({
  getNotificationStatusApi: () => mockGetStatus(),
}));

describe('notificationStore', () => {
  beforeEach(() => {
    useNotificationStore.setState({ unreadCount: 0 });
    mockGetStatus.mockReset();
  });

  it('starts with zero unread', () => {
    expect(useNotificationStore.getState().unreadCount).toBe(0);
  });

  it('fetchUnreadCount updates from API', async () => {
    mockGetStatus.mockResolvedValue({ data: { unreadCount: 5 } });
    await useNotificationStore.getState().fetchUnreadCount();
    expect(useNotificationStore.getState().unreadCount).toBe(5);
  });

  it('setUnreadCount sets the count', () => {
    useNotificationStore.getState().setUnreadCount(3);
    expect(useNotificationStore.getState().unreadCount).toBe(3);
  });
});