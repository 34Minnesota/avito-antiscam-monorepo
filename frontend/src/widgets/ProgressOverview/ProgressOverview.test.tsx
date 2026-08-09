import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ProgressOverview } from './ProgressOverview';
import { Progress } from '@/shared/api/contracts';

const base: Progress = {
  totalScenarios: 2,
  completedScenarios: 0,
  passedScenarios: 0,
  completionPercent: 0,
  passedPercent: 0,
  roles: [
    {
      role: 'buyer',
      totalScenarios: 2,
      completedScenarios: 0,
      passedScenarios: 0,
      completionPercent: 0,
      passedPercent: 0,
      scenarios: [],
    },
    {
      role: 'seller',
      totalScenarios: 2,
      completedScenarios: 0,
      passedScenarios: 0,
      completionPercent: 0,
      passedPercent: 0,
      scenarios: [],
    },
  ],
  roleComparison: { completionPercentDelta: 0, passedPercentDelta: 0 },
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
            { ...base.roles[1], completedScenarios: 1, completionPercent: 50 },
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
            { ...base.roles[0], completedScenarios: 1, completionPercent: 50 },
            { ...base.roles[1], completedScenarios: 1, completionPercent: 75 },
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
              completedScenarios: 1,
              completionPercent: 100,
              scenarios: [
                {
                  scenarioSlug: 'safe-deal',
                  title: 'Срочная предоплата',
                  completed: true,
                  passed: true,
                  attemptsCount: 2,
                  recentAttempts: [
                    {
                      attemptId: 'a2',
                      score: { points: 95, maxPoints: 100, percent: 95 },
                      outcome: 'safe',
                      completedAt: '2026-08-09T10:00:00Z',
                    },
                    {
                      attemptId: 'a1',
                      score: { points: 70, maxPoints: 100, percent: 70 },
                      outcome: 'partial',
                      completedAt: '2026-08-08T10:00:00Z',
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
