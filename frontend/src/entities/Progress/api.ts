import { Progress } from '@/shared/api/contracts';
import { rtkApi } from '@/shared/api/rtkApi';

export const progressApi = rtkApi.injectEndpoints({
  endpoints: (build) => ({
    getProgress: build.query<Progress, void>({
      query: () => 'v1/progress',
      providesTags: ['Progress'],
    }),
  }),
});

export const { useGetProgressQuery } = progressApi;
