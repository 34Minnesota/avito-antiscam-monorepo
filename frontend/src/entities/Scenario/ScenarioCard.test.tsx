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
          scenario_slug: 'safe-deal',
          title: scenario.title,
          completed: false,
          passed: false,
          attempts_count: 1,
          active_attempt_id: 'a1',
          recent_attempts: [],
        }}
        onStart={onStart}
      />,
    );
    expect(screen.getByRole('button', { name: 'Продолжить' })).toBeInTheDocument();
    expect(screen.getByText(/Лучший результат/)).toBeInTheDocument();

    rerender(
      <ScenarioCard
        scenario={scenario}
        progress={{
          scenario_slug: 'safe-deal',
          title: scenario.title,
          completed: true,
          passed: true,
          attempts_count: 2,
          active_attempt_id: null,
          latest_score: { points: 80, max_points: 100, percent: 80 },
          recent_attempts: [],
        }}
        onStart={onStart}
      />,
    );
    expect(screen.getByRole('button', { name: 'Пройти ещё раз' })).toBeInTheDocument();
  });
});
