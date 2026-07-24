import request from '@/utils/request';
import type { ApiResponse, Segment, PaginatedResponse, ListParams, SegmentCondition } from '@/types';

export const getSegmentsApi = (params?: ListParams): Promise<ApiResponse<PaginatedResponse<Segment>>> =>
  request({ url: '/manage/segments', method: 'get', params });

export const getSegmentDetailsApi = (id: string): Promise<ApiResponse<Segment>> =>
  request({ url: `/manage/segments/${id}`, method: 'get' });

export const createSegmentApi = (data: { name: string; description?: string; matchType?: string; conditions?: SegmentCondition[] }): Promise<ApiResponse<Segment>> =>
  request({ url: '/manage/segments', method: 'post', data });

export const updateSegmentApi = (id: string, data: Partial<{ name: string; description: string; matchType: string; conditions: SegmentCondition[] }>): Promise<ApiResponse<Segment>> =>
  request({ url: `/manage/segments/${id}`, method: 'put', data });

export const deleteSegmentApi = (id: string): Promise<ApiResponse<null>> =>
  request({ url: `/manage/segments/${id}`, method: 'delete' });