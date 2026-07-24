import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import FeatureFlags from './FeatureFlags';
import * as featureFlagsApi from '@/api/featureFlags';

vi.mock('@/api/featureFlags', () => ({
  getFeatureFlagsApi: vi.fn(),
  getFeatureFlagDetailsApi: vi.fn(),
  createFeatureFlagApi: vi.fn(),
  updateFeatureFlagApi: vi.fn(),
  deleteFeatureFlagApi: vi.fn(),
  getFeatureFlagHistoryApi: vi.fn(),
  archiveFeatureFlagApi: vi.fn(),
  unarchiveFeatureFlagApi: vi.fn(),
}));

vi.mock('@/api/segments', () => ({
  getSegmentsApi: vi.fn(() => Promise.resolve({ data: { items: [] } })),
}));

vi.mock('@/utils/formatters', () => ({
  formatDateTime: vi.fn(() => '2026-01-01 00:00'),
  formatFlagType: vi.fn((t) => t),
}));

let mockCurrentEnvId: string | null = 'env-1';

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useOutletContext: () => ({ currentEnvId: mockCurrentEnvId }),
  };
});

describe('FeatureFlags', () => {
  afterEach(() => {
    vi.clearAllMocks();
    mockCurrentEnvId = 'env-1';
  });

  it('renders title and create button', () => {
    vi.mocked(featureFlagsApi.getFeatureFlagsApi).mockResolvedValue({ code: 200, message: 'ok', data: { items: [], total: 0, page: 1, pageSize: 10, totalPages: 0 } });

    render(
      <BrowserRouter>
        <FeatureFlags />
      </BrowserRouter>
    );
    expect(screen.getByText('Feature Flags')).toBeTruthy();
    expect(screen.getByText('Create')).toBeTruthy();
  });

  it('shows info alert when no environment selected', () => {
    mockCurrentEnvId = null;

    render(
      <BrowserRouter>
        <FeatureFlags />
      </BrowserRouter>
    );
    expect(screen.getByText('Select an environment to manage feature flags.')).toBeTruthy();
  });

  it('renders flags table when data is available', async () => {
    vi.mocked(featureFlagsApi.getFeatureFlagsApi).mockResolvedValue({
      code: 200,
      message: 'ok',
      data: {
        items: [
          { id: 'ff-1', name: 'Test Flag', flagType: 'RELEASE', enabled: true, rolloutPercentage: 50, expiresAt: 'never', createdAt: '2026-01-01', updatedAt: '2026-01-01' },
        ],
        total: 1,
        page: 1,
        pageSize: 10,
        totalPages: 1,
      },
    });

    render(
      <BrowserRouter>
        <FeatureFlags />
      </BrowserRouter>
    );

    expect(await screen.findByText('Test Flag')).toBeTruthy();
    expect(screen.getByText('Enabled')).toBeTruthy();
    expect(screen.getByText('50%')).toBeTruthy();
  });

  it('shows empty message when no flags', async () => {
    vi.mocked(featureFlagsApi.getFeatureFlagsApi).mockResolvedValue({ code: 200, message: 'ok', data: { items: [], total: 0, page: 1, pageSize: 10, totalPages: 0 } });

    render(
      <BrowserRouter>
        <FeatureFlags />
      </BrowserRouter>
    );

    expect(await screen.findByText('No feature flags yet')).toBeTruthy();
  });

  it('shows loading spinner while fetching', () => {
    vi.mocked(featureFlagsApi.getFeatureFlagsApi).mockReturnValue(new Promise(() => {}));

    render(
      <BrowserRouter>
        <FeatureFlags />
      </BrowserRouter>
    );

    expect(document.querySelector('.MuiCircularProgress-root')).toBeTruthy();
  });

  it('shows error alert on fetch failure', async () => {
    vi.mocked(featureFlagsApi.getFeatureFlagsApi).mockRejectedValue(new Error('Network error'));

    render(
      <BrowserRouter>
        <FeatureFlags />
      </BrowserRouter>
    );

    expect(await screen.findByText('Network error')).toBeTruthy();
  });
});
