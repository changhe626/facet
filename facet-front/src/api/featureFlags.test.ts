import { describe, it, expect, vi } from 'vitest';
import request from '@/utils/request';
import { getFeatureFlagsApi, getFeatureFlagDetailsApi, createFeatureFlagApi, updateFeatureFlagApi, deleteFeatureFlagApi } from './featureFlags';

vi.mock('@/utils/request');
const mockRequest = vi.mocked(request);

describe('featureFlags API', () => {
  it('getFeatureFlagsApi calls GET with envId', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { items: [] } });
    await getFeatureFlagsApi('env-1', { page: 1 });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/environments/env-1/features',
      method: 'get',
      params: { page: 1 },
    });
  });

  it('getFeatureFlagDetailsApi calls GET with envId and flagId', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: {} });
    await getFeatureFlagDetailsApi('env-1', 'ff-1');
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/environments/env-1/features/ff-1',
      method: 'get',
    });
  });

  it('createFeatureFlagApi calls POST', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 'ff-1' } });
    await createFeatureFlagApi('env-1', { name: 'Test' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/environments/env-1/features',
      method: 'post',
      data: { name: 'Test' },
    });
  });

  it('updateFeatureFlagApi calls PUT', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: null as never });
    await updateFeatureFlagApi('env-1', 'ff-1', { enabled: true });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/environments/env-1/features/ff-1',
      method: 'put',
      data: { enabled: true },
    });
  });

  it('deleteFeatureFlagApi calls DELETE', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: null as never });
    await deleteFeatureFlagApi('env-1', 'ff-1');
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/environments/env-1/features/ff-1',
      method: 'delete',
    });
  });
});