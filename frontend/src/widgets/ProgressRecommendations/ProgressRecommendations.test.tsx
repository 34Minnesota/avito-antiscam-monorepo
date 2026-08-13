import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ProgressRecommendation, ScenarioCard } from '@/shared/api/contracts';
import { ProgressRecommendations } from './ProgressRecommendations';

const scenarios: ScenarioCard[] = [
  {
    id: 'buyer-1',
    slug: 'buyer-1',
    role: 'buyer',
    category: 'Оплата',
    difficulty: 1,
    title: 'Покупка 1',
    description: 'Описание',
  },
  {
    id: 'buyer-2',
    slug: 'buyer-2',
    role: 'buyer',
    category: 'Оплата',
    difficulty: 1,
    title: 'Покупка 2',
    description: 'Описание',
  },
  {
    id: 'buyer-3',
    slug: 'buyer-3',
    role: 'buyer',
    category: 'Оплата',
    difficulty: 1,
    title: 'Покупка 3',
    description: 'Описание',
  },
  {
    id: 'buyer-4',
    slug: 'buyer-4',
    role: 'buyer',
    category: 'Оплата',
    difficulty: 1,
    title: 'Покупка 4',
    description: 'Описание',
  },
  {
    id: 'seller-1',
    slug: 'seller-1',
    role: 'seller',
    category: 'Доставка',
    difficulty: 1,
    title: 'Продажа 1',
    description: 'Описание',
  },
];

const recommendations: ProgressRecommendation[] = [
  {
    role: 'buyer',
    scenario_slug: 'buyer-1',
    reason_code: 'ACTIVE_ATTEMPT',
    reason_text: 'Есть активная попытка.',
  },
  {
    role: 'buyer',
    scenario_slug: 'buyer-2',
    reason_code: 'REPEAT_FOR_REINFORCEMENT',
    reason_text: 'Закрепите навык.',
  },
  {
    role: 'buyer',
    scenario_slug: 'buyer-3',
    reason_code: 'NOT_STARTED',
    reason_text: 'Сценарий не начат.',
  },
  {
    role: 'buyer',
    scenario_slug: 'buyer-4',
    reason_code: 'NOT_STARTED',
    reason_text: 'Лишняя рекомендация.',
  },
  {
    role: 'seller',
    scenario_slug: 'seller-1',
    reason_code: 'NOT_STARTED',
    reason_text: 'Рекомендация продавца.',
  },
];

describe('ProgressRecommendations', () => {
  it('shows three recommendations for the active role with titles and reasons', () => {
    render(
      <ProgressRecommendations
        role="buyer"
        recommendations={recommendations}
        scenarios={scenarios}
        onStart={vi.fn()}
      />,
    );

    expect(
      screen.getByRole('heading', { name: 'Рекомендации для покупателя' }),
    ).toBeInTheDocument();
    expect(screen.getAllByRole('listitem')).toHaveLength(3);
    expect(screen.getByText('Покупка 1')).toBeInTheDocument();
    expect(screen.getByText('Есть активная попытка.')).toBeInTheDocument();
    expect(screen.queryByText('Покупка 4')).not.toBeInTheDocument();
    expect(screen.queryByText('Продажа 1')).not.toBeInTheDocument();
  });

  it('uses status-aware actions and starts the selected scenario', () => {
    const onStart = vi.fn();
    render(
      <ProgressRecommendations
        role="buyer"
        recommendations={recommendations}
        scenarios={scenarios}
        onStart={onStart}
      />,
    );

    const cards = screen.getAllByRole('listitem');
    expect(within(cards[0]).getByRole('button', { name: 'Продолжить' })).toBeInTheDocument();
    expect(within(cards[1]).getByRole('button', { name: 'Пройти ещё раз' })).toBeInTheDocument();
    fireEvent.click(within(cards[2]).getByRole('button', { name: 'Начать' }));
    expect(onStart).toHaveBeenCalledWith(scenarios[2]);
  });
});
