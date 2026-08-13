import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { TrainingChat } from './TrainingChat';

describe('TrainingChat', () => {
  it('shows an accessible counterpart typing indicator', () => {
    render(<TrainingChat messages={[]} counterpartName="Алексей" isCounterpartTyping />);

    expect(screen.getByRole('status', { name: 'Алексей печатает' })).toBeVisible();
  });

  it('does not show the indicator when the counterpart is idle', () => {
    render(<TrainingChat messages={[]} counterpartName="Алексей" />);

    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });
});
