import { describe, expect, it } from 'vitest';
import { Progress, ScenarioCard } from '@/shared/api/contracts';
import { getExperience, getProgressDynamics, getSkills } from './insights';

const progress: Progress = {
  total_scenarios: 2,
  completed_scenarios: 1,
  passed_scenarios: 1,
  completion_percent: 50,
  passed_percent: 100,
  roles: [
    {
      role: 'seller',
      total_scenarios: 2,
      completed_scenarios: 1,
      passed_scenarios: 1,
      completion_percent: 50,
      passed_percent: 100,
      scenarios: [
        {
          scenario_slug: 'delivery',
          title: 'Доставка',
          completed: true,
          passed: true,
          attempts_count: 2,
          best_score: { points: 90, max_points: 100, percent: 90 },
          initial_score: { points: 65, max_points: 100, percent: 65 },
          latest_score: { points: 90, max_points: 100, percent: 90 },
          improvement_percent_points: 25,
          trend: 'improving',
          recent_attempts: [],
        },
        {
          scenario_slug: 'payment',
          title: 'Оплата',
          completed: false,
          passed: false,
          attempts_count: 0,
          recent_attempts: [],
        },
      ],
    },
    {
      role: 'buyer',
      total_scenarios: 0,
      completed_scenarios: 0,
      passed_scenarios: 0,
      completion_percent: 0,
      passed_percent: 0,
      scenarios: [],
    },
  ],
  role_comparison: { completion_percent_delta: 0, passed_percent_delta: 0 },
  recommendations: [],
  experience: {
    total_xp: 165,
    level: 1,
    current_xp: 165,
    next_level_xp: 250,
    achievements: [],
  },
};

const catalog: ScenarioCard[] = [
  {
    id: '1',
    slug: 'delivery',
    role: 'seller',
    category: 'fake_delivery',
    difficulty: 2,
    title: 'Доставка',
    description: '',
  },
  {
    id: '2',
    slug: 'payment',
    role: 'seller',
    category: 'Оплата',
    difficulty: 3,
    title: 'Оплата',
    description: '',
  },
];

describe('progress insights', () => {
  it('uses server-calculated XP and exposes role completion counts', () => {
    expect(getExperience(progress, 'seller')).toEqual({
      xp: 165,
      level: 1,
      currentLevelXp: 165,
      nextLevelXp: 250,
      progressPercent: 66,
      completedScenarios: 1,
      safeScenarios: 1,
    });
  });

  it('builds skills from completed scenario categories', () => {
    expect(getSkills(progress, 'seller', catalog)).toEqual([
      { name: 'Проверка доставки', score: 90, completed: 1 },
    ]);
  });

  it('calculates improvement dynamics from first to latest result', () => {
    expect(getProgressDynamics(progress, 'seller')).toEqual({
      initial: 65,
      latest: 90,
      delta: 25,
      trend: 'improving',
      scenariosTracked: 1,
    });
  });

  it('derives dynamics from recent attempts when explicit initial/latest fields are absent', () => {
    const historyOnly: Progress = {
      ...progress,
      roles: [
        {
          ...progress.roles[0],
          scenarios: [
            {
              ...progress.roles[0].scenarios[0],
              initial_score: null,
              latest_score: null,
              best_score: null,
              recent_attempts: [
                {
                  attempt_id: 'old',
                  score: { points: 30, max_points: 100, percent: 30 },
                  outcome: 'partial',
                  completed_at: '2026-08-08T10:00:00Z',
                },
                {
                  attempt_id: 'new',
                  score: { points: 80, max_points: 100, percent: 80 },
                  outcome: 'safe',
                  completed_at: '2026-08-09T10:00:00Z',
                },
              ],
            },
          ],
        },
        progress.roles[1],
      ],
    };

    expect(getProgressDynamics(historyOnly, 'seller')).toEqual({
      initial: 30,
      latest: 80,
      delta: 50,
      trend: 'improving',
      scenariosTracked: 1,
    });
  });

  it('returns an empty dynamics state before a scenario has a first result', () => {
    expect(getProgressDynamics(progress, 'buyer')).toEqual({
      initial: 0,
      latest: 0,
      delta: 0,
      trend: 'none',
      scenariosTracked: 0,
    });
  });
});
