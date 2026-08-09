import { rtkApi } from '@/shared/api/rtkApi';
import { Role, ScenarioCard } from '@/shared/api/contracts';
export const scenarioApi = rtkApi.injectEndpoints({
  endpoints: (build) => ({
    getScenarios: build.query<{ scenarios: ScenarioCard[] }, Role | undefined>({
      query: (role) => ({ url: 'v1/scenarios', params: role ? { role } : undefined }),
      providesTags: ['Scenarios'],
    }),
  }),
});
export const { useGetScenariosQuery } = scenarioApi;
