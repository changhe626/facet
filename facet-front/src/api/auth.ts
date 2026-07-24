import request from '@/utils/request';
import type { ApiResponse, LoginResponse } from '@/types';

export const getCaptchaApi = (): Promise<ApiResponse<{ captchaId: string; base64Image: string }>> =>
  request({ url: '/manage/auth/captcha', method: 'get' });

export const loginApi = (data: { email: string; password: string; captchaId?: string; captchaAnswer?: string }): Promise<ApiResponse<LoginResponse>> =>
  request({ url: '/manage/auth/login', method: 'post', data });