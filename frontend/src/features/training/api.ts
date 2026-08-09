import { rtkApi } from '@/shared/api/rtkApi';
import { ChoiceResult, StartResult, Summary } from '@/shared/api/contracts';

export const trainingApi = rtkApi.injectEndpoints({
  endpoints: (build) => ({
    startAttempt: build.mutation<StartResult, string>({
      query: (scenarioId) => ({
        url: 'v1/attempts',
        method: 'POST',
        body: { scenario_id: scenarioId },
      }),
      invalidatesTags: [],
    }),
    submitChoice: build.mutation<
      ChoiceResult,
      { attemptId: string; sceneId: string; optionId: string; expectedRevision: number }
    >({
      query: ({ attemptId, sceneId, optionId, expectedRevision }) => ({
        url: `v1/attempts/${attemptId}/choice`,
        method: 'POST',
        body: { scene_id: sceneId, option_id: optionId, expected_revision: expectedRevision },
      }),
      invalidatesTags: (result) => (result?.finished ? ['Progress', 'Scenarios'] : []),
    }),
    getSummary: build.query<Summary, string>({
      query: (attemptId) => `v1/attempts/${attemptId}/summary`,
      providesTags: (_result, _error, attemptId) => [{ type: 'Summary', id: attemptId }],
    }),
  }),
});

export const { useStartAttemptMutation, useSubmitChoiceMutation, useGetSummaryQuery } = trainingApi;
