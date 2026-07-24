import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Users from './Users';

vi.mock('@/api/user', () => ({
  getUsersApi: () => Promise.resolve({ data: { items: [], total: 0 } }),
}));

describe('Users', () => {
  it('renders the page title', async () => {
    render(<BrowserRouter><Users /></BrowserRouter>);
    expect(screen.getByText('Users')).toBeTruthy();
  });
});