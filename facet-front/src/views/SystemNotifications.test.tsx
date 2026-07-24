import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import SystemNotifications from './SystemNotifications';

vi.mock('@/api/systemNotifications', () => ({
  getSystemNotificationsApi: () => Promise.resolve({ data: { items: [], total: 0 } }),
}));

describe('SystemNotifications', () => {
  it('renders the page title', async () => {
    render(<SystemNotifications />);
    expect(screen.getByText('System Notifications')).toBeTruthy();
  });
});