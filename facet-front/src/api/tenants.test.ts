import { describe, it, expect, vi } from 'vitest';
import request from '@/utils/request';
import { getTenantsApi, getTenantDetailsApi, createTenantApi, updateTenantApi, resetTenantPasswordApi } from './tenants';

vi.mock('@/utils/request');
const mockRequest = vi.mocked(request);

describe('tenants API', () => {
  it('getTenantsApi calls GET with params', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { items: [] } });
    await getTenantsApi({ page: 1, pageSize: 10 });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/system/tenants',
      method: 'get',
      params: { page: 1, pageSize: 10 },
    });
  });

  it('getTenantDetailsApi calls GET with id', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 't-1' } });
    await getTenantDetailsApi('t-1');
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/system/tenants/t-1',
      method: 'get',
    });
  });

  it('createTenantApi calls POST with data', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 't-1' } });
    await createTenantApi({
      tenant_info: { name: 'Test' },
      user_info: { role: 'tenant_admin', adminEmail: 'a@b.com', adminInitialPassword: 'pwd' },
    });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/system/tenants',
      method: 'post',
      data: {
        tenant_info: { name: 'Test' },
        user_info: { role: 'tenant_admin', adminEmail: 'a@b.com', adminInitialPassword: 'pwd' },
      },
    });
  });

  it('updateTenantApi calls PUT with id', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: { id: 't-1' } });
    await updateTenantApi('t-1', { name: 'Updated' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/system/tenants/t-1',
      method: 'put',
      data: { name: 'Updated' },
    });
  });

  it('resetTenantPasswordApi calls PUT /password', async () => {
    mockRequest.mockResolvedValue({ code: 200, message: 'OK', data: null });
    await resetTenantPasswordApi('t-1', { new_password: 'newpwd' });
    expect(mockRequest).toHaveBeenCalledWith({
      url: '/system/tenants/t-1/password',
      method: 'put',
      data: { new_password: 'newpwd' },
    });
  });
});
