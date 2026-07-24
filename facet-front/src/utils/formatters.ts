import { useLocaleStore } from '@/stores/localeStore';

const getT = () => useLocaleStore.getState().t;

export const formatRole = (role: string): string => {
  const t = getT();
  const roleMap: Record<string, string> = {
    tenant_admin: t.systemTenants.roleReadWrite,
    tenant_viewer: t.systemTenants.roleReadOnly,
    tenant_super_admin: t.systemTenants.roleAdmin,
    system_admin: 'System Admin',
  };
  return roleMap[role] || role;
};

export const formatDateTime = (isoString?: string | null): string => {
  if (!isoString) return '-';
  const d = new Date(isoString);
  if (isNaN(d.getTime()) || d.getFullYear() <= 1) return '-';
  const y = d.getFullYear();
  const mo = String(d.getMonth() + 1).padStart(2, '0');
  const da = String(d.getDate()).padStart(2, '0');
  const h = String(d.getHours()).padStart(2, '0');
  const mi = String(d.getMinutes()).padStart(2, '0');
  return `${y}-${mo}-${da} ${h}:${mi}`;
};

export const formatFlagType = (type: string): string => {
  const map: Record<string, string> = {
    RELEASE: 'Release',
    EXPERIMENT: 'Experiment',
    OPERATIONAL: 'Operational',
    PERMISSION: 'Permission',
  };
  return map[type] || type;
};
