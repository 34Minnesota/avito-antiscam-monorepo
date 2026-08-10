import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import { apiUrl } from './base';
import { getSessionId } from '@/shared/auth/session';

export const rtkApi = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    baseUrl: apiUrl('/'),
    prepareHeaders: (headers) => {
      const sessionId = getSessionId();
      if (sessionId) headers.set('X-Session-ID', sessionId);
      return headers;
    },
  }),
  tagTypes: ['User', 'Progress', 'Scenarios', 'Summary'],
  endpoints: () => ({}),
});
