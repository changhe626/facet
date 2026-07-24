import request from '@/utils/request';
import type { ApiResponse, User, PaginatedResponse, ListParams, CreateUserRequest, UpdateUserRequest } from '@/types';

export const getUsersApi = (params?: ListParams): Promise<ApiResponse<PaginatedResponse<User>>> =>
  request({ url: '/manage/users', method: 'get', params });

export const createUserApi = (data: CreateUserRequest): Promise<ApiResponse<User>> =>
  request({ url: '/manage/users', method: 'post', data });

export const updateUserApi = (userId: string, data: UpdateUserRequest): Promise<ApiResponse<User>> =>
  request({ url: `/manage/users/${userId}`, method: 'put', data });

export const deleteUserApi = (userId: string): Promise<ApiResponse<null>> =>
  request({ url: `/manage/users/${userId}`, method: 'delete' });

export const getUserDetailsApi = (userId: string): Promise<ApiResponse<User>> =>
  request({ url: `/manage/users/${userId}`, method: 'get' });

export const resetUserPasswordApi = (userId: string): Promise<ApiResponse<null>> =>
  request({ url: `/manage/users/${userId}/password`, method: 'put' });