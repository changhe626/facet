import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import HelpCenter from './HelpCenter';

describe('HelpCenter', () => {
  it('renders help sections', () => {
    render(<HelpCenter />);
    expect(screen.getByText('Help Center')).toBeTruthy();
    expect(screen.getByText('Multi-Tenant Architecture')).toBeTruthy();
    expect(screen.getByText('Feature Flags')).toBeTruthy();
    expect(screen.getByText('Getting Started')).toBeTruthy();
    expect(screen.getByText('Best Practices')).toBeTruthy();
  });
});