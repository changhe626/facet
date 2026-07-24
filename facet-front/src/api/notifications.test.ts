import { describe, it, expect, vi } from 'vitest';
import request from '@/utils/request';
import { getNotificationStatusApi, getNotificationsApi, markNotificationsAsReadApi } from './notifications';

vi.mock('@/utils/request');
const mockRequest = vi.mocked(request);

describe('notifications API', () => {
  it('getNotificationStatusApi calls GET', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { unreadCount: 3 } });
    await getNotificationStatusApi();
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/notifications/status',
      method: 'get',
    });
  });

  it('getNotificationsApi calls GET', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: [] });
    await getNotificationsApi();
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/notifications',
      method: 'get',
    });
  });

  it('markNotificationsAsReadApi calls POST with ids', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: null });
    await markNotificationsAsReadApi(['n1', 'n2']);
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/notifications/mark-as-read',
      method: 'post',
      data: { notificationIds: ['n1', 'n2'] },
    });
  });
});
