import { describe, expect, it } from 'vitest';
import { getProgressHeadline } from './progress';

describe('getProgressHeadline', () => {
  it('prioritizes an active attempt', () => {
    expect(
      getProgressHeadline({
        completed: false,
        activeAttemptId: 'attempt-1',
        improvementPercentPoints: null,
        latestScore: null,
      }),
    ).toBe('Тренировка не завершена');
  });
  it('shows an explicit not-started state', () => {
    expect(
      getProgressHeadline({
        completed: false,
        activeAttemptId: null,
        improvementPercentPoints: null,
        latestScore: null,
      }),
    ).toBe('Ещё не пройден');
  });
  it('shows improvement before the latest score', () => {
    expect(
      getProgressHeadline({
        completed: true,
        activeAttemptId: null,
        improvementPercentPoints: 18,
        latestScore: { points: 72, maxPoints: 100, percent: 72 },
      }),
    ).toBe('+18 баллов к первому результату');
  });
  it('falls back to the latest score', () => {
    expect(
      getProgressHeadline({
        completed: true,
        activeAttemptId: null,
        improvementPercentPoints: 0,
        latestScore: { points: 72, maxPoints: 100, percent: 72 },
      }),
    ).toBe('72% — последний результат');
  });
});
