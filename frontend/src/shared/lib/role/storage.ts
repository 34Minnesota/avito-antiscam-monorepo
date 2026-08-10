import { ROLE_KEY } from '@/shared/const/localstorage';
import { Role } from '@/shared/api/contracts';

export const getStoredRole = (): Role | null => {
  const value = localStorage.getItem(ROLE_KEY);
  return value === 'buyer' || value === 'seller' ? value : null;
};

export const saveStoredRole = (role: Role) => localStorage.setItem(ROLE_KEY, role);
