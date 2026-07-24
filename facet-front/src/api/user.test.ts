import { describe, it, expect, vi } from 'vitest';
import request from '@/utils/request';
import { getUsersApi, createUserApi, updateUserApi, deleteUserApi, getUserDetailsApi, resetUserPasswordApi } from './user';

vi.mock('@/utils/request');
const mockRequest = vi.mocked(request);

describe('user API', () => {
  it('getUsersApi calls GET with params', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { items: [] } });
    await getUsersApi({ page: 1, pageSize: 10 });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/users',
      method: 'get',
      params: { page: 1, pageSize: 10 },
    });
  });

  it('createUserApi calls POST with data', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 'u-1' } });
    await createUserApi({ email: 'a@b.com', role: 'admin', initialPassword: 'pwd123' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/users',
      method: 'post',
      data: { email: 'a@b.com', role: 'admin', initialPassword: 'pwd123' },
    });
  });

  it('updateUserApi calls PUT with id', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 'u-1' } });
    await updateUserApi('u-1', { email: 'a@b.com', role: 'viewer' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/users/u-1',
      method: 'put',
      data: { email: 'a@b.com', role: 'viewer' },
    });
  });

  it('deleteUserApi calls DELETE with id', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: null });
    await deleteUserApi('u-1');
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/users/u-1',
      method: 'delete',
    });
  });

  it('getUserDetailsApi calls GET with id', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 'u-1' } });
    await getUserDetailsApi('u-1');
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/users/u-1',
      method: 'get',
    });
  });

  it('resetUserPasswordApi calls PUT /password', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: null });
    await resetUserPasswordApi('u-1');
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/manage/users/u-1/password',
      method: 'put',
    });
  });
});
