import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Segments from './Segments';
import * as segmentsApi from '@/api/segments';

vi.mock('@/api/segments', () => ({
  getSegmentsApi: vi.fn(),
  createSegmentApi: vi.fn(),
  updateSegmentApi: vi.fn(),
  deleteSegmentApi: vi.fn(),
}));

vi.mock('@/utils/formatters', () => ({
  formatDateTime: vi.fn(() => '2026-01-01 00:00'),
}));

describe('Segments', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders title and create button', () => {
    vi.mocked(segmentsApi.getSegmentsApi).mockResolvedValue({ code: 200, message: 'ok', data: { items: [], total: 0, page: 1, pageSize: 10, totalPages: 0 } });

    render(
      <BrowserRouter>
        <Segments />
      </BrowserRouter>
    );
    expect(screen.getByText('Segments')).toBeTruthy();
    expect(screen.getByText('Create')).toBeTruthy();
  });

  it('renders segments table when data is available', async () => {
    vi.mocked(segmentsApi.getSegmentsApi).mockResolvedValue({
      code: 200,
      message: 'ok',
      data: {
        items: [
          { id: 'seg-1', name: 'VIP Users', matchType: 'ANY', conditions: [{ attributeKey: 'tier', operator: 'equals', value: 'vip' }], createdAt: '2026-01-01', updatedAt: '2026-01-01' },
        ],
        total: 1,
        page: 1,
        pageSize: 10,
        totalPages: 1,
      },
    });

    render(
      <BrowserRouter>
        <Segments />
      </BrowserRouter>
    );

    expect(await screen.findByText('VIP Users')).toBeTruthy();
    expect(screen.getByText('1 condition(s)')).toBeTruthy();
  });

  it('shows empty message when no segments', async () => {
    vi.mocked(segmentsApi.getSegmentsApi).mockResolvedValue({ code: 200, message: 'ok', data: { items: [], total: 0, page: 1, pageSize: 10, totalPages: 0 } });

    render(
      <BrowserRouter>
        <Segments />
      </BrowserRouter>
    );

    expect(await screen.findByText('No segments yet')).toBeTruthy();
  });

  it('shows loading spinner while fetching', () => {
    vi.mocked(segmentsApi.getSegmentsApi).mockReturnValue(new Promise(() => {}));

    render(
      <BrowserRouter>
        <Segments />
      </BrowserRouter>
    );

    expect(document.querySelector('.MuiCircularProgress-root')).toBeTruthy();
  });

  it('shows error alert on fetch failure', async () => {
    vi.mocked(segmentsApi.getSegmentsApi).mockRejectedValue(new Error('Network error'));

    render(
      <BrowserRouter>
        <Segments />
      </BrowserRouter>
    );

    expect(await screen.findByText('Network error')).toBeTruthy();
  });
});
