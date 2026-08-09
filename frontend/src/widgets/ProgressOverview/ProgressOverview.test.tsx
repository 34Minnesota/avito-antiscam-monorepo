import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ProgressOverview } from './ProgressOverview';
import { Progress } from '@/shared/api/contracts';

const base: Progress = {
  total_scenarios: 2,
  completed_scenarios: 0,
  passed_scenarios: 0,
  completion_percent: 0,
  passed_percent: 0,
  roles: [
    {
      role: 'buyer',
      total_scenarios: 2,
      completed_scenarios: 0,
      passed_scenarios: 0,
      completion_percent: 0,
      passed_percent: 0,
      scenarios: [],
    },
    {
      role: 'seller',
      total_scenarios: 2,
      completed_scenarios: 0,
      passed_scenarios: 0,
      completion_percent: 0,
      passed_percent: 0,
      scenarios: [],
    },
  ],
  role_comparison: { completion_percent_delta: 0, passed_percent_delta: 0 },
  recommendations: [],
};

describe('ProgressOverview', () => {
  it('does not show misleading -100 percentage points', () => {
    render(
      <ProgressOverview
        progress={{
          ...base,
          roles: [
            base.roles[0],
            { ...base.roles[1], completed_scenarios: 1, completion_percent: 50 },
          ],
        }}
        activeRole="seller"
      />,
    );
    expect(screen.queryByText(/баллов/)).not.toBeInTheDocument();
    expect(screen.getByText('Сравнение появится позже')).toBeInTheDocument();
  });

  it('shows a comparison only after both roles have a completed scenario', () => {
    render(
      <ProgressOverview
        progress={{
          ...base,
          roles: [
            { ...base.roles[0], completed_scenarios: 1, completion_percent: 50 },
            { ...base.roles[1], completed_scenarios: 1, completion_percent: 75 },
          ],
        }}
        activeRole="seller"
      />,
    );
    expect(screen.getByText('Продавец выше на 25 баллов')).toBeInTheDocument();
    expect(screen.getByText('Продавец 75% · Покупатель 50%')).toBeInTheDocument();
  });

  it('shows the latest attempts as a compact learning history', () => {
    render(
      <ProgressOverview
        progress={{
          ...base,
          roles: [
            { ...base.roles[0] },
            {
              ...base.roles[1],
              completed_scenarios: 1,
              completion_percent: 100,
              scenarios: [
                {
                  scenario_slug: 'safe-deal',
                  title: 'Срочная предоплата',
                  completed: true,
                  passed: true,
                  attempts_count: 2,
                  recent_attempts: [
                    {
                      attempt_id: 'a2',
                      score: { points: 95, max_points: 100, percent: 95 },
                      outcome: 'safe',
                      completed_at: '2026-08-09T10:00:00Z',
                    },
                    {
                      attempt_id: 'a1',
                      score: { points: 70, max_points: 100, percent: 70 },
                      outcome: 'partial',
                      completed_at: '2026-08-08T10:00:00Z',
                    },
                  ],
                },
              ],
            },
          ],
        }}
        activeRole="seller"
      />,
    );

    expect(screen.getByText('Ваш прогресс в динамике')).toBeInTheDocument();
    expect(screen.getByText('95%')).toBeInTheDocument();
    expect(screen.getByText('70%')).toBeInTheDocument();
    expect(screen.getAllByText('Срочная предоплата')).toHaveLength(2);
  });

  it('explains the empty history state', () => {
    render(<ProgressOverview progress={base} activeRole="seller" />);
    expect(screen.getByText(/Завершите первую тренировку/)).toBeInTheDocument();
  });
});
