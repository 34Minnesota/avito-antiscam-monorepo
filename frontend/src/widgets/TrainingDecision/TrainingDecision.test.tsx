import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { TrainingDecision } from './TrainingDecision';

describe('TrainingDecision', () => {
  const options = [
    { id: 'safe', text: 'Проверить данные' },
    { id: 'risky', text: 'Перевести деньги' },
  ];

  it('renders all decisions with accessible buttons', () => {
    render(<TrainingDecision prompt="Что сделать?" options={options} onChoose={vi.fn()} />);
    expect(screen.getByRole('button', { name: /Проверить данные/ })).toBeEnabled();
    expect(screen.getByRole('button', { name: /Перевести деньги/ })).toBeEnabled();
  });

  it('sends the selected option and disables choices while saving', async () => {
    const user = userEvent.setup();
    const onChoose = vi.fn();
    const { rerender } = render(
      <TrainingDecision prompt="Что сделать?" options={options} onChoose={onChoose} disabled />,
    );
    const button = screen.getByRole('button', { name: /Проверить данные/ });
    expect(button).toBeDisabled();

    rerender(<TrainingDecision prompt="Что сделать?" options={options} onChoose={onChoose} />);
    await user.click(screen.getByRole('button', { name: /Проверить данные/ }));
    expect(onChoose).toHaveBeenCalledWith('safe', 'Проверить данные');
  });
});
