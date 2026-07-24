import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import Overview from './Overview';

describe('Overview', () => {
  it('renders welcome message and feature cards', () => {
    render(<Overview />);
    expect(screen.getByText('Overview')).toBeTruthy();
    expect(screen.getAllByText('Feature Flags').length).toBeGreaterThan(0);
    expect(screen.getByText('A/B Testing')).toBeTruthy();
    expect(screen.getByText('Multi-Tenant')).toBeTruthy();
    expect(screen.getByText('Real-time Evaluation')).toBeTruthy();
  });
});