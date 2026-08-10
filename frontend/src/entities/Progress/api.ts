import { Progress } from '@/shared/api/contracts';
import { rtkApi } from '@/shared/api/rtkApi';
import { validateProgress } from '@/shared/api/validation';

export const progressApi = rtkApi.injectEndpoints({
  endpoints: (build) => ({
    getProgress: build.query<Progress, void>({
      transformResponse: validateProgress,
      query: () => 'v1/progress',
      providesTags: ['Progress'],
    }),
  }),
});

export const { useGetProgressQuery } = progressApi;
