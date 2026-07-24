import { describe, it, expect, vi } from 'vitest';
import request from '@/utils/request';
import { getSegmentsApi, getSegmentDetailsApi, createSegmentApi, updateSegmentApi, deleteSegmentApi } from './segments';

vi.mock('@/utils/request');
const mockRequest = vi.mocked(request);

describe('segments API', () => {
  it('getSegmentsApi calls GET with params', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { items: [] } });
    await getSegmentsApi({ page: 1, pageSize: 10 });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/segments',
      method: 'get',
      params: { page: 1, pageSize: 10 },
    });
  });

  it('getSegmentDetailsApi calls GET with id', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 'seg-1' } });
    await getSegmentDetailsApi('seg-1');
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/segments/seg-1',
      method: 'get',
    });
  });

  it('createSegmentApi calls POST with data', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 'seg-1' } });
    await createSegmentApi({ name: 'Test Segment' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/segments',
      method: 'post',
      data: { name: 'Test Segment' },
    });
  });

  it('updateSegmentApi calls PUT with id and data', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 'seg-1' } });
    await updateSegmentApi('seg-1', { name: 'Updated' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/segments/seg-1',
      method: 'put',
      data: { name: 'Updated' },
    });
  });

  it('deleteSegmentApi calls DELETE with id', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: null });
    await deleteSegmentApi('seg-1');
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/segments/seg-1',
      method: 'delete',
    });
  });
});
