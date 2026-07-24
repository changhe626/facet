import { describe, it, expect, vi } from 'vitest';
import request from '@/utils/request';
import { loginApi, getCaptchaApi } from './auth';

vi.mock('@/utils/request');
const mockRequest = vi.mocked(request);

describe('auth API', () => {
  it('loginApi calls POST', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { token: 't', user: {} } });
    await loginApi({ email: 'a@b.com', password: 'pwd' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/auth/login',
      method: 'post',
      data: { email: 'a@b.com', password: 'pwd' },
    });
  });

  it('getCaptchaApi calls GET', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { base64Image: 'img' } });
    await getCaptchaApi();
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/auth/captcha',
      method: 'get',
    });
  });
});