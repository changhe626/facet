import request from '@/utils/request';
import type { ApiResponse, Notification, NotificationStatus } from '@/types';

export const getNotificationStatusApi = (): Promise<ApiResponse<NotificationStatus>> =>
  request({ url: '/manage/notifications/status', method: 'get' });

export const getNotificationsApi = (): Promise<ApiResponse<Notification[]>> =>
  request({ url: '/manage/notifications', method: 'get' });

export const markNotificationsAsReadApi = (notificationIds: string[]): Promise<ApiResponse<null>> =>
  request({ url: '/manage/notifications/mark-as-read', method: 'post', data: { notificationIds } });