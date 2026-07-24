import { describe, it, expect } from 'vitest';
import { formatRole, formatDateTime, formatFlagType } from './formatters';

describe('formatRole', () => {
  it('returns label for known roles', () => {
    expect(formatRole('tenant_admin')).toBe('Read & Write');
    expect(formatRole('tenant_viewer')).toBe('Read Only');
    expect(formatRole('tenant_super_admin')).toBe('Admin');
    expect(formatRole('system_admin')).toBe('System Admin');
  });

  it('returns the input for unknown roles', () => {
    expect(formatRole('custom_role')).toBe('custom_role');
    expect(formatRole('')).toBe('');
  });
});

describe('formatDateTime', () => {
  it('formats a valid ISO string', () => {
    const result = formatDateTime('2025-10-03T10:00:00Z');
    expect(result).not.toBe('-');
    expect(result).not.toBe('');
  });

  it('returns "-" for null', () => {
    expect(formatDateTime(null)).toBe('-');
  });

  it('returns "-" for undefined', () => {
    expect(formatDateTime(undefined)).toBe('-');
  });

  it('returns "-" for zero date', () => {
    expect(formatDateTime('0001-01-01T00:00:00Z')).toBe('-');
  });
});

describe('formatFlagType', () => {
  it('returns label for known types', () => {
    expect(formatFlagType('RELEASE')).toBe('Release');
    expect(formatFlagType('EXPERIMENT')).toBe('Experiment');
    expect(formatFlagType('OPERATIONAL')).toBe('Operational');
    expect(formatFlagType('PERMISSION')).toBe('Permission');
  });

  it('returns the input for unknown types', () => {
    expect(formatFlagType('CUSTOM')).toBe('CUSTOM');
  });
});
