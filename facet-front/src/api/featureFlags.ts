import request from '@/utils/request';
import type { ApiResponse, FeatureFlag, PaginatedResponse, ListParams, CreateFeatureFlagRequest, UpdateFeatureFlagRequest, FeatureFlagDetail, FeatureFlagHistory } from '@/types';

export const getFeatureFlagsApi = (envId: string, params?: ListParams): Promise<ApiResponse<PaginatedResponse<FeatureFlag>>> =>
  request({ url: `/manage/environments/${envId}/features`, method: 'get', params });

export const getFeatureFlagDetailsApi = (envId: string, ffId: string): Promise<ApiResponse<FeatureFlagDetail>> =>
  request({ url: `/manage/environments/${envId}/features/${ffId}`, method: 'get' });

export const getFeatureFlagHistoryApi = (envId: string, ffId: string): Promise<ApiResponse<FeatureFlagHistory[]>> =>
  request({ url: `/manage/environments/${envId}/features/${ffId}/history`, method: 'get' });

export const createFeatureFlagApi = (envId: string, data: CreateFeatureFlagRequest): Promise<ApiResponse<FeatureFlag>> =>
  request({ url: `/manage/environments/${envId}/features`, method: 'post', data });

export const updateFeatureFlagApi = (envId: string, ffId: string, data: UpdateFeatureFlagRequest): Promise<ApiResponse<FeatureFlag>> =>
  request({ url: `/manage/environments/${envId}/features/${ffId}`, method: 'put', data });

export const deleteFeatureFlagApi = (envId: string, ffId: string): Promise<ApiResponse<null>> =>
  request({ url: `/manage/environments/${envId}/features/${ffId}`, method: 'delete' });

export const archiveFeatureFlagApi = (envId: string, ffId: string): Promise<ApiResponse<null>> =>
  request({ url: `/manage/environments/${envId}/features/${ffId}/archive`, method: 'post' });

export const unarchiveFeatureFlagApi = (envId: string, ffId: string): Promise<ApiResponse<null>> =>
  request({ url: `/manage/environments/${envId}/features/${ffId}/unarchive`, method: 'post' });