import { rtkApi } from '@/shared/api/rtkApi';
import { SessionResponse, User } from '@/shared/api/contracts';
export const userApi = rtkApi.injectEndpoints({
  endpoints: (build) => ({
    getMe: build.query<User, void>({ query: () => 'v1/users/me', providesTags: ['User'] }),
    login: build.mutation<SessionResponse, { email: string; password: string }>({
      query: (body) => ({ url: 'v1/auth/login', method: 'POST', body }),
    }),
    register: build.mutation<
      SessionResponse,
      { nickname: string; email: string; password: string }
    >({ query: (body) => ({ url: 'v1/auth/register', method: 'POST', body }) }),
    logout: build.mutation<void, void>({
      query: () => ({ url: 'v1/auth/logout', method: 'POST' }),
      invalidatesTags: ['User', 'Progress', 'Scenarios'],
    }),
  }),
});
export const { useGetMeQuery, useLoginMutation, useRegisterMutation, useLogoutMutation } = userApi;
