import { http, HttpResponse } from 'msw';

const API_BASE = '/api/v1';

export const handlers = [
  http.post(`${API_BASE}/manage/auth/login`, () =>
    HttpResponse.json({
      code: 200,
      message: 'OK',
      data: { token: 'mock-token', user: { id: 'u-1', email: 'admin@facet.com', role: 'system_admin' } },
    })
  ),
  http.get(`${API_BASE}/manage/auth/captcha`, () =>
    HttpResponse.json({
      code: 200,
      message: 'OK',
      data: { captchaId: 'mock-captcha', base64Image: '' },
    })
  ),
  http.get(`${API_BASE}/manage/environments`, () =>
    HttpResponse.json({
      code: 200,
      message: 'OK',
      data: { items: [], total: 0, page: 1, pageSize: 10, totalPages: 0 },
    })
  ),
];
