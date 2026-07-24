import axios, { type AxiosRequestConfig } from 'axios';
import { STORAGE_KEYS, API_BASE_URL, AXIOS_TIMEOUT } from '@/utils/constants';
import { navigate } from '@/utils/navigate';
import type { ApiResponse } from '@/types';

const service = axios.create({
  baseURL: API_BASE_URL,
  timeout: AXIOS_TIMEOUT,
});

service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem(STORAGE_KEYS.TOKEN);
    if (token && config.headers) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

service.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response) {
      const { status, data } = error.response;
      const msg = data?.message || error.message;
      if (status === 401) {
        localStorage.removeItem(STORAGE_KEYS.TOKEN);
        localStorage.removeItem(STORAGE_KEYS.USER);
        if (window.location.pathname !== '/login') {
          navigate('/login', { replace: true });
        }
        return Promise.reject(new Error(msg));
      }
      return Promise.reject(new Error(msg));
    }
    return Promise.reject(new Error(error.message || 'Network error'));
  }
);

function request<T = unknown>(config: AxiosRequestConfig): Promise<ApiResponse<T>> {
  return service.request(config) as Promise<ApiResponse<T>>;
}

request.get = <T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> =>
  service.get(url, config) as Promise<ApiResponse<T>>;

request.post = <T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> =>
  service.post(url, data, config) as Promise<ApiResponse<T>>;

request.put = <T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> =>
  service.put(url, data, config) as Promise<ApiResponse<T>>;

request.delete = <T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> =>
  service.delete(url, config) as Promise<ApiResponse<T>>;

export default request;
