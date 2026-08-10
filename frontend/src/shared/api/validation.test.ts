import { describe, expect, it } from 'vitest';
import { validateProgress } from './validation';

describe('validateProgress', () => {
  it('normalizes backend camelCase role metrics and keeps real progress', () => {
    const result = validateProgress({
      total_scenarios: 8,
      completed_scenarios: 4,
      passed_scenarios: 3,
      completion_percent: 50,
      passed_percent: 38,
      roles: [
        {
          role: 'seller',
          totalScenarios: 8,
          completedScenarios: 4,
          passedScenarios: 3,
          completionPercent: 50,
          passedPercent: 38,
          scenarios: [],
        },
      ],
      role_comparison: { completion_percent_delta: 0, passed_percent_delta: 0 },
      recommendations: [],
      experience: {
        total_xp: 240,
        level: 2,
        current_xp: 40,
        next_level_xp: 100,
        achievements: [],
      },
    });

    expect(result.roles[0]).toMatchObject({
      total_scenarios: 8,
      completed_scenarios: 4,
      passed_scenarios: 3,
      completion_percent: 50,
      passed_percent: 38,
    });
  });

  it('derives counts when aggregate counters are absent instead of producing NaN/0', () => {
    const result = validateProgress({
      roles: [
        {
          role: 'seller',
          scenarios: [
            { scenario_slug: 'a', title: 'A', completed: true, passed: true, recent_attempts: [] },
            { scenario_slug: 'b', title: 'B', completed: true, passed: false, recent_attempts: [] },
            {
              scenario_slug: 'c',
              title: 'C',
              completed: false,
              passed: false,
              recent_attempts: [],
            },
          ],
        },
      ],
      experience: {},
      recommendations: [],
    });

    expect(result.roles[0].total_scenarios).toBe(3);
    expect(result.roles[0].completed_scenarios).toBe(2);
    expect(result.roles[0].passed_scenarios).toBe(1);
    expect(result.roles[0].completion_percent).toBe(67);
    expect(result.roles[0].passed_percent).toBe(33);
  });

  it('keeps the recommendation role', () => {
    const result = validateProgress({
      roles: [],
      recommendations: [
        {
          role: 'buyer',
          scenario_slug: 'buyer-scenario',
          reason_code: 'NOT_STARTED',
          reason_text: 'Начните сценарий.',
        },
      ],
      experience: {},
    });

    expect(result.recommendations[0]).toMatchObject({
      role: 'buyer',
      scenario_slug: 'buyer-scenario',
    });
  });

  it('rejects a recommendation with an invalid role', () => {
    expect(() =>
      validateProgress({
        roles: [],
        recommendations: [
          {
            role: 'admin',
            scenario_slug: 'scenario',
            reason_code: 'NOT_STARTED',
            reason_text: 'Начните сценарий.',
          },
        ],
        experience: {},
      }),
    ).toThrow('Invalid API response: ProgressRecommendation.role');
  });
});
