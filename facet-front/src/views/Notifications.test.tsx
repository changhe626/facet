import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import Notifications from './Notifications';

vi.mock('@/api/notifications', () => ({
  getNotificationsApi: () => Promise.resolve({ data: [] }),
}));

describe('Notifications', () => {
  it('renders the page title', async () => {
    render(<Notifications />);
    expect(screen.getByText('Notifications')).toBeTruthy();
  });
});