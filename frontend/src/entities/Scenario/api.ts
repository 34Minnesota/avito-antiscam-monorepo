import { rtkApi } from '@/shared/api/rtkApi';
import { validateScenarioList } from '@/shared/api/validation';
import { Role, ScenarioCard } from '@/shared/api/contracts';
export const scenarioApi = rtkApi.injectEndpoints({
  endpoints: (build) => ({
    getScenarios: build.query<{ scenarios: ScenarioCard[] }, Role | undefined>({
      transformResponse: validateScenarioList,
      query: (role) => ({ url: 'v1/scenarios', params: role ? { role } : undefined }),
      providesTags: ['Scenarios'],
    }),
  }),
});
export const { useGetScenariosQuery } = scenarioApi;
