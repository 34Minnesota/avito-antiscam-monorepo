import { SESSION_KEY } from '@/shared/const/localstorage';

export const getSessionId = () => localStorage.getItem(SESSION_KEY);
export const setSessionId = (id: string) => localStorage.setItem(SESSION_KEY, id);
export const clearSessionId = () => localStorage.removeItem(SESSION_KEY);
