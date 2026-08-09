import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ScenarioCard } from './ScenarioCard';

const scenario = {
  id: '1',
  slug: 'safe-deal',
  role: 'seller' as const,
  category: 'Оплата',
  difficulty: 3,
  title: 'Срочная предоплата',
  description: 'Тест',
  stats: { best_score: 70, attempts_count: 1 },
};

describe('ScenarioCard', () => {
  it('distinguishes an active attempt from a completed scenario', () => {
    const onStart = vi.fn();
    const { rerender } = render(
      <ScenarioCard
        scenario={scenario}
        progress={{
          scenarioSlug: 'safe-deal',
          title: scenario.title,
          completed: false,
          passed: false,
          attemptsCount: 1,
          activeAttemptId: 'a1',
          recentAttempts: [],
        }}
        onStart={onStart}
      />,
    );
    expect(screen.getByRole('button', { name: 'Продолжить' })).toBeInTheDocument();

    rerender(
      <ScenarioCard
        scenario={scenario}
        progress={{
          scenarioSlug: 'safe-deal',
          title: scenario.title,
          completed: true,
          passed: true,
          attemptsCount: 2,
          activeAttemptId: null,
          latestScore: { points: 80, maxPoints: 100, percent: 80 },
          recentAttempts: [],
        }}
        onStart={onStart}
      />,
    );
    expect(screen.getByRole('button', { name: 'Пройти ещё раз' })).toBeInTheDocument();
  });
});
