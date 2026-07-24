import { useAuthStore } from '@/stores/authStore';

const WRITE_ROLES = ['system_admin', 'tenant_super_admin', 'tenant_admin'];

export function useCanWrite() {
  const role = useAuthStore((s) => s.user?.role);
  return !!role && WRITE_ROLES.includes(role);
}
