import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import Environments from './Environments';

vi.mock('@/api/environment', () => ({
  getEnvironmentsApi: () => Promise.resolve({ data: { items: [], total: 0 } }),
}));

describe('Environments', () => {
  it('renders the page title', async () => {
    render(<Environments />);
    expect(screen.getByText('Environments')).toBeTruthy();
  });
});