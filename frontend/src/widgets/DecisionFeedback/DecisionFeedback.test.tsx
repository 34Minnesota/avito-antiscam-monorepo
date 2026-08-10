import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { DecisionFeedback } from './DecisionFeedback';

describe('DecisionFeedback', () => {
  it('announces the verdict and exposes the final navigation action', async () => {
    const user = userEvent.setup();
    const onContinue = vi.fn();
    render(
      <DecisionFeedback
        verdict="safe"
        text="Проверка была правильной."
        isFinal
        onContinue={onContinue}
      />,
    );
    expect(screen.getByRole('region')).toHaveAttribute('aria-live', 'polite');
    expect(screen.getByText('Безопасное решение')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Посмотреть итог' }));
    expect(onContinue).toHaveBeenCalledOnce();
  });
});
