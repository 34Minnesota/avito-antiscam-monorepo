import { RoleProgress } from '@/shared/api/contracts';

export type RoleComparison =
  | { state: 'empty'; value: string; note: string }
  | { state: 'partial'; value: string; note: string }
  | { state: 'equal'; value: string; note: string }
  | { state: 'ready'; value: string; note: string };

export const getRoleComparison = (roles: RoleProgress[]): RoleComparison => {
  const buyer = roles.find((role) => role.role === 'buyer');
  const seller = roles.find((role) => role.role === 'seller');
  const buyerStarted = Boolean(buyer?.completed_scenarios);
  const sellerStarted = Boolean(seller?.completed_scenarios);

  if (!buyerStarted && !sellerStarted)
    return { state: 'empty', value: 'Пока нет данных', note: 'Пройдите хотя бы одну тренировку' };
  if (!buyerStarted || !sellerStarted)
    return {
      state: 'partial',
      value: 'Сравнение появится позже',
      note: 'Пройдите хотя бы одну тренировку в обеих ролях',
    };

  const delta = (seller?.completion_percent ?? 0) - (buyer?.completion_percent ?? 0);
  const sellerPercent = seller?.completion_percent ?? 0;
  const buyerPercent = buyer?.completion_percent ?? 0;
  const note = `Продавец ${sellerPercent}% · Покупатель ${buyerPercent}%`;

  if (delta === 0) return { state: 'equal', value: 'Результаты равны', note };

  const leader = delta > 0 ? 'Продавец' : 'Покупатель';
  return { state: 'ready', value: `${leader} выше на ${Math.abs(delta)} баллов`, note };
};
