import request from '@/utils/request';
import type { ApiResponse, Environment, PaginatedResponse, ListParams, CreateEnvironmentRequest, UpdateEnvironmentRequest } from '@/types';

export const getActiveEnvironmentsApi = (): Promise<ApiResponse<Environment[]>> =>
  request({ url: '/manage/environments/active-list', method: 'get' });

export const getEnvironmentsApi = (params?: ListParams): Promise<ApiResponse<PaginatedResponse<Environment>>> =>
  request({ url: '/manage/environments', method: 'get', params });

export const createEnvironmentApi = (data: CreateEnvironmentRequest): Promise<ApiResponse<Environment>> =>
  request({ url: '/manage/environments', method: 'post', data });

export const updateEnvironmentApi = (envId: string, data: UpdateEnvironmentRequest): Promise<ApiResponse<Environment>> =>
  request({ url: `/manage/environments/${envId}`, method: 'put', data });

export const deleteEnvironmentApi = (envId: string): Promise<ApiResponse<null>> =>
  request({ url: `/manage/environments/${envId}`, method: 'delete' });