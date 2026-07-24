import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import SystemTenants from './SystemTenants';

vi.mock('@/api/tenants', () => ({
  getTenantsApi: () => Promise.resolve({ data: { items: [], total: 0 } }),
}));

describe('SystemTenants', () => {
  it('renders the page title', async () => {
    render(<SystemTenants />);
    expect(screen.getByText('Tenants')).toBeTruthy();
  });
});