import request from '@/utils/request';
import type { ApiResponse, Tenant, PaginatedResponse, ListParams } from '@/types';

export const getTenantsApi = (params?: ListParams): Promise<ApiResponse<PaginatedResponse<Tenant>>> =>
  request({ url: '/system/tenants', method: 'get', params });

export const getTenantDetailsApi = (tenantId: string): Promise<ApiResponse<Tenant>> =>
  request({ url: `/system/tenants/${tenantId}`, method: 'get' });

export const createTenantApi = (data: {
  tenant_info: { name: string; levelId?: string; description?: string; website?: string; status?: string };
  user_info: { role: string; adminEmail: string; adminDisplayName?: string; adminInitialPassword: string };
}): Promise<ApiResponse<Tenant>> =>
  request({ url: '/system/tenants', method: 'post', data });

export const updateTenantApi = (tenantId: string, data: Partial<{ name: string; levelId: string; description: string; website: string; status: string }>): Promise<ApiResponse<Tenant>> =>
  request({ url: `/system/tenants/${tenantId}`, method: 'put', data });

export const resetTenantPasswordApi = (tenantId: string, data: { new_password: string }): Promise<ApiResponse<null>> =>
  request({ url: `/system/tenants/${tenantId}/password`, method: 'put', data });