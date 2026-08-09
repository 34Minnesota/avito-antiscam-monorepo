import { ScenarioCard as ScenarioCardType, ScenarioProgress } from '@/shared/api/contracts';
import { Badge } from '@/shared/ui/Badge';
import { Button } from '@/shared/ui/Button';
import { Card } from '@/shared/ui/Card';
import { getProgressHeadline } from '@/shared/lib/progress/progress';
import cls from './ScenarioCard.module.scss';

const difficulty = (value: number) =>
  `${'●'.repeat(Math.min(value, 5))}${'○'.repeat(Math.max(0, 5 - value))}`;

interface Props {
  scenario: ScenarioCardType;
  progress?: ScenarioProgress;
  onStart: () => void;
}

export const ScenarioCard = ({ scenario, progress, onStart }: Props) => {
  const active = Boolean(progress?.active_attempt_id);
  const completed = Boolean(progress?.completed);
  const buttonLabel = active ? 'Продолжить' : completed ? 'Пройти ещё раз' : 'Начать';
  const headline = progress
    ? getProgressHeadline(progress)
    : scenario.stats
      ? `${scenario.stats.attempts_count} ${scenario.stats.attempts_count === 1 ? 'попытка' : 'попыток'}`
      : 'Начните с первой попытки';

  return (
    <Card className={cls.Card}>
      <div className={cls.top}>
        <Badge tone={scenario.role === 'seller' ? 'accent' : 'good'}>
          {scenario.role === 'seller' ? 'Продавец' : 'Покупатель'}
        </Badge>
        <span className={cls.difficulty} aria-label={`Сложность ${scenario.difficulty} из 5`}>
          {difficulty(scenario.difficulty)}
        </span>
      </div>
      <div className={cls.icon} aria-hidden="true">
        {scenario.role === 'seller' ? '↗' : '↙'}
      </div>
      <h3>{scenario.title}</h3>
      <p>{scenario.description}</p>
      <div className={cls.meta}>
        <span>{scenario.category}</span>
        <span>{progress?.attempts_count ?? scenario.stats?.attempts_count ?? 0} попыток</span>
      </div>
      <div className={cls.bottom}>
        <strong>{headline}</strong>
        <Button type="button" onClick={onStart}>
          {buttonLabel}
        </Button>
      </div>
    </Card>
  );
};
