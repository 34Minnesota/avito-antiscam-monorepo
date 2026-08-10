import { rtkApi } from '@/shared/api/rtkApi';
import { SessionResponse, User } from '@/shared/api/contracts';
import { validateUser } from '@/shared/api/validation';

export const userApi = rtkApi.injectEndpoints({
  endpoints: (build) => ({
    getMe: build.query<User, void>({
      transformResponse: validateUser,
      query: () => 'v1/users/me',
      providesTags: ['User'],
    }),
    login: build.mutation<SessionResponse, { email: string; password: string }>({
      query: (body) => ({ url: 'v1/auth/login', method: 'POST', body }),
    }),
    register: build.mutation<
      SessionResponse,
      { nickname: string; email: string; password: string }
    >({
      query: (body) => ({ url: 'v1/auth/register', method: 'POST', body }),
    }),
    logout: build.mutation<void, void>({
      query: () => ({ url: 'v1/auth/logout', method: 'POST' }),
      invalidatesTags: ['User', 'Progress', 'Scenarios'],
    }),
  }),
});
export const { useGetMeQuery, useLoginMutation, useRegisterMutation, useLogoutMutation } = userApi;
