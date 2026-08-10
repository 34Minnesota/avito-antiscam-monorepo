import { API_URL } from '@/shared/config/env';

export const apiUrl = (path: string) => `${API_URL}${path}`;
