import { describe, it, expect, vi } from 'vitest';
import request from '@/utils/request';
import { getSystemNotificationsApi, createSystemNotificationApi, getSystemNotificationDetailsApi, updateSystemNotificationApi, deleteSystemNotificationApi } from './systemNotifications';

vi.mock('@/utils/request');
const mockRequest = vi.mocked(request);

describe('systemNotifications API', () => {
  it('getSystemNotificationsApi calls GET with params', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { items: [] } });
    await getSystemNotificationsApi({ page: 1, pageSize: 10 });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/system/notifications',
      method: 'get',
      params: { page: 1, pageSize: 10 },
    });
  });

  it('createSystemNotificationApi calls POST', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 'sn-1' } });
    await createSystemNotificationApi({ title: 'Test', content: 'Hello', level: 'info', type: 'broadcast' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/system/notifications',
      method: 'post',
      data: { title: 'Test', content: 'Hello', level: 'info', type: 'broadcast' },
    });
  });

  it('getSystemNotificationDetailsApi calls GET', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 'sn-1' } });
    await getSystemNotificationDetailsApi('sn-1');
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/system/notifications/sn-1',
      method: 'get',
    });
  });

  it('updateSystemNotificationApi calls PUT', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 'sn-1' } });
    await updateSystemNotificationApi('sn-1', { title: 'Updated', content: 'New', level: 'info', type: 'broadcast' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/system/notifications/sn-1',
      method: 'put',
      data: { title: 'Updated', content: 'New', level: 'info', type: 'broadcast' },
    });
  });

  it('deleteSystemNotificationApi calls DELETE', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: null });
    await deleteSystemNotificationApi('sn-1');
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/system/notifications/sn-1',
      method: 'delete',
    });
  });
});
