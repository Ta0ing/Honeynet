import axios from 'axios';

export const api = axios.create({ baseURL: '/api/v1', timeout: 15_000 });

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('honeynet_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  (response) => response.data.data,
  (error) => {
    const publicBrandingRequest = ['/platform/branding', '/platform/oem'].includes(error.config?.url || '');
    if (!publicBrandingRequest && error.response?.status === 401 && !location.pathname.includes('/login')) {
      localStorage.removeItem('honeynet_token');
      localStorage.removeItem('honeynet_user');
      location.href = '/login';
    }
    return Promise.reject(error);
  },
);

export const errorMessage = (error: unknown) => {
  if (axios.isAxiosError(error)) return error.response?.data?.message || error.message;
  return error instanceof Error ? error.message : '请求失败';
};

export const formatTime = (value?: string) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—';
