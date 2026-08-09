import { describe, expect, it } from 'vitest';
import { getRoleComparison } from './comparison';
import { RoleProgress } from '@/shared/api/contracts';

const role = (
  name: 'buyer' | 'seller',
  completedScenarios: number,
  completionPercent: number,
): RoleProgress => ({
  role: name,
  totalScenarios: 2,
  completedScenarios,
  passedScenarios: completedScenarios,
  completionPercent,
  passedPercent: completionPercent,
  scenarios: [],
});

describe('getRoleComparison', () => {
  it('does not manufacture a negative comparison before any training', () => {
    expect(getRoleComparison([role('buyer', 0, 0), role('seller', 0, 0)])).toEqual({
      state: 'empty',
      value: 'Пока нет данных',
      note: 'Пройдите хотя бы одну тренировку',
    });
  });

  it('asks for both roles before comparing them', () => {
    expect(getRoleComparison([role('buyer', 0, 0), role('seller', 1, 50)]).state).toBe('partial');
  });

  it('shows equal results explicitly', () => {
    expect(getRoleComparison([role('buyer', 1, 50), role('seller', 1, 50)])).toMatchObject({
      state: 'equal',
      value: 'Результаты равны',
    });
  });

  it('explains the difference between completed progress', () => {
    expect(getRoleComparison([role('buyer', 1, 40), role('seller', 1, 70)])).toMatchObject({
      state: 'ready',
      value: 'Продавец выше на 30 баллов',
    });
    expect(getRoleComparison([role('buyer', 1, 70), role('seller', 1, 40)])).toMatchObject({
      state: 'ready',
      value: 'Покупатель выше на 30 баллов',
    });
  });
});
