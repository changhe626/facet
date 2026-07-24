import request from '@/utils/request';
import type { ApiResponse, SystemNotification, PaginatedResponse, ListParams, CreateSystemNotificationRequest } from '@/types';

export const getSystemNotificationsApi = (params?: ListParams): Promise<ApiResponse<PaginatedResponse<SystemNotification>>> =>
  request({ url: '/system/notifications', method: 'get', params });

export const createSystemNotificationApi = (data: CreateSystemNotificationRequest): Promise<ApiResponse<SystemNotification>> =>
  request({ url: '/system/notifications', method: 'post', data });

export const getSystemNotificationDetailsApi = (notificationId: string): Promise<ApiResponse<SystemNotification>> =>
  request({ url: `/system/notifications/${notificationId}`, method: 'get' });

export const updateSystemNotificationApi = (notificationId: string, data: CreateSystemNotificationRequest): Promise<ApiResponse<SystemNotification>> =>
  request({ url: `/system/notifications/${notificationId}`, method: 'put', data });

export const deleteSystemNotificationApi = (notificationId: string): Promise<ApiResponse<null>> =>
  request({ url: `/system/notifications/${notificationId}`, method: 'delete' });