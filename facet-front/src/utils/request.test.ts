import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockUse = vi.fn();
const mockInstance = {
  interceptors: {
    request: { use: mockUse },
    response: { use: mockUse },
  },
};

vi.mock('axios', () => ({
  default: { create: vi.fn(() => mockInstance) },
  create: vi.fn(() => mockInstance),
}));

describe('request service', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('creates axios instance with /api/v1 baseURL and registers interceptors', async () => {
    const axios = (await import('axios')).default;
    await import('./request');

    expect(axios.create).toHaveBeenCalledWith(
      expect.objectContaining({ baseURL: '/api/v1', timeout: 5000 })
    );
    expect(mockUse).toHaveBeenCalledTimes(2);
  });
});
