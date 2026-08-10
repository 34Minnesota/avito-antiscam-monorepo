import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Progress } from '@/shared/api/contracts';
import { ProgressInsights } from './ProgressInsights';

const progress: Progress = {
  total_scenarios: 1,
  completed_scenarios: 1,
  passed_scenarios: 1,
  completion_percent: 100,
  passed_percent: 100,
  roles: [
    {
      role: 'seller',
      total_scenarios: 1,
      completed_scenarios: 1,
      passed_scenarios: 1,
      completion_percent: 100,
      passed_percent: 100,
      scenarios: [
        {
          scenario_slug: 'payment',
          title: 'Оплата',
          completed: true,
          passed: true,
          attempts_count: 2,
          best_score: { points: 90, max_points: 100, percent: 90 },
          initial_score: { points: 60, max_points: 100, percent: 60 },
          latest_score: { points: 90, max_points: 100, percent: 90 },
          improvement_percent_points: 30,
          trend: 'improving',
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
  role_comparison: { completion_percent_delta: 100, passed_percent_delta: 100 },
  recommendations: [],
  experience: {
    total_xp: 165,
    level: 1,
    current_xp: 165,
    next_level_xp: 250,
    achievements: [
      {
        code: 'FIRST_COMPLETION',
        title: 'Первый шаг',
        description: 'Завершён первый сценарий.',
        earned: true,
      },
      {
        code: 'FIRST_SAFE_RESULT',
        title: 'Безопасный исход',
        description: 'Получен первый безопасный результат.',
        earned: true,
      },
      {
        code: 'BOTH_ROLES',
        title: 'Две роли',
        description: 'Есть успешно пройденные сценарии покупателя и продавца.',
        earned: false,
      },
    ],
  },
};

describe('ProgressInsights', () => {
  it('shows level, XP, skills and improvement dynamics', () => {
    render(
      <ProgressInsights
        progress={progress}
        role="seller"
        scenarios={[
          {
            id: '1',
            slug: 'payment',
            role: 'seller',
            category: 'Оплата',
            difficulty: 3,
            title: 'Оплата',
            description: 'Сценарий',
          },
        ]}
      />,
    );

    expect(screen.getByText('УРОВЕНЬ')).toBeInTheDocument();
    expect(screen.getByText('165 XP')).toBeInTheDocument();
    expect(screen.getByText('Первый шаг')).toBeInTheDocument();
    expect(screen.getByText('2 / 3')).toBeInTheDocument();
    expect(screen.getByText('Две роли')).toBeInTheDocument();
    expect(screen.getByText('ДИНАМИКА РЕЗУЛЬТАТА')).toBeInTheDocument();
    expect(screen.getByText('+30')).toBeInTheDocument();
    expect(screen.getByText('Оплата')).toBeInTheDocument();
    expect(screen.getByText('90%')).toBeInTheDocument();
  });

  it('explains what happens before the first completed scenario', () => {
    const empty = {
      ...progress,
      completed_scenarios: 0,
      passed_scenarios: 0,
      roles: progress.roles.map((item) =>
        item.role === 'seller'
          ? {
              ...item,
              completed_scenarios: 0,
              passed_scenarios: 0,
              completion_percent: 0,
              passed_percent: 0,
              scenarios: [],
            }
          : item,
      ),
    };
    render(<ProgressInsights progress={empty} role="seller" scenarios={[]} />);

    expect(screen.getByText('Навыки появятся после первого результата.')).toBeInTheDocument();
    expect(screen.getByText(/Пройдите сценарий повторно/)).toBeInTheDocument();
  });
});
