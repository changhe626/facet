import { describe, it, expect, vi } from 'vitest';
import request from '@/utils/request';
import { getEnvironmentsApi, createEnvironmentApi, updateEnvironmentApi, deleteEnvironmentApi, getActiveEnvironmentsApi } from './environment';

vi.mock('@/utils/request');
const mockRequest = vi.mocked(request);

describe('environment API', () => {
  it('getEnvironmentsApi calls GET with params', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { items: [] } });
    await getEnvironmentsApi({ page: 1, pageSize: 10 });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/environments',
      method: 'get',
      params: { page: 1, pageSize: 10 },
    });
  });

  it('createEnvironmentApi calls POST with data', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 'env-1' } });
    await createEnvironmentApi({ name: 'Test' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/environments',
      method: 'post',
      data: { name: 'Test' },
    });
  });

  it('updateEnvironmentApi calls PUT', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: null as never });
    await updateEnvironmentApi('env-1', { name: 'Updated' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/environments/env-1',
      method: 'put',
      data: { name: 'Updated' },
    });
  });

  it('deleteEnvironmentApi calls DELETE', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: null as never });
    await deleteEnvironmentApi('env-1');
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/environments/env-1',
      method: 'delete',
    });
  });

  it('getActiveEnvironmentsApi calls GET /manage/environments/active-list', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: [] });
    await getActiveEnvironmentsApi();
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/environments/active-list',
      method: 'get',
    });
  });
});