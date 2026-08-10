import { describe, expect, it } from 'vitest';
import { getProgressHeadline } from './progress';

describe('getProgressHeadline', () => {
  it('prioritizes an active attempt', () => {
    expect(
      getProgressHeadline({
        completed: false,
        active_attempt_id: 'attempt-1',
        improvement_percent_points: null,
        latest_score: null,
      }),
    ).toBe('Тренировка не завершена');
  });
  it('shows an explicit not-started state', () => {
    expect(
      getProgressHeadline({
        completed: false,
        active_attempt_id: null,
        improvement_percent_points: null,
        latest_score: null,
      }),
    ).toBe('Ещё не пройден');
  });
  it('shows improvement before the latest score', () => {
    expect(
      getProgressHeadline({
        completed: true,
        active_attempt_id: null,
        improvement_percent_points: 18,
        latest_score: { points: 72, max_points: 100, percent: 72 },
      }),
    ).toBe('+18 баллов к первому результату');
  });
  it('falls back to the latest score', () => {
    expect(
      getProgressHeadline({
        completed: true,
        active_attempt_id: null,
        improvement_percent_points: 0,
        latest_score: { points: 72, max_points: 100, percent: 72 },
      }),
    ).toBe('72% — последний результат');
  });
});
