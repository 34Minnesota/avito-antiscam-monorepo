import {
  ProgressRecommendation,
  Role,
  ScenarioCard as ScenarioCardType,
} from '@/shared/api/contracts';
import { Button } from '@/shared/ui/Button';
import { Card } from '@/shared/ui/Card';
import cls from './ProgressRecommendations.module.scss';

interface Props {
  role: Role;
  recommendations: ProgressRecommendation[];
  scenarios: ScenarioCardType[];
  onStart: (scenario: ScenarioCardType) => void;
}

const actionLabel = (reasonCode: string) => {
  if (reasonCode === 'ACTIVE_ATTEMPT') return 'Продолжить';
  if (reasonCode === 'REPEAT_FOR_REINFORCEMENT') return 'Пройти ещё раз';
  return 'Начать';
};

export const ProgressRecommendations = ({ role, recommendations, scenarios, onStart }: Props) => {
  const items = recommendations
    .filter((recommendation) => recommendation.role === role)
    .slice(0, 3)
    .flatMap((recommendation) => {
      const scenario = scenarios.find((item) => item.slug === recommendation.scenario_slug);
      return scenario ? [{ recommendation, scenario }] : [];
    });

  if (items.length === 0) return null;

  const roleLabel = role === 'seller' ? 'продавца' : 'покупателя';

  return (
    <section className={cls.section} aria-labelledby="recommendations-title">
      <div className={cls.header}>
        <div>
          <span>СЛЕДУЮЩИЕ ШАГИ</span>
          <h2 id="recommendations-title">Рекомендации для {roleLabel}</h2>
        </div>
        <p>Сценарии подобраны по вашему текущему прогрессу.</p>
      </div>

      <div className={cls.grid} role="list">
        {items.map(({ recommendation, scenario }, index) => (
          <Card className={cls.card} role="listitem" key={recommendation.scenario_slug}>
            <span className={cls.position}>РЕКОМЕНДАЦИЯ {index + 1}</span>
            <h3>{scenario.title}</h3>
            <p>{recommendation.reason_text}</p>
            <Button type="button" onClick={() => onStart(scenario)}>
              {actionLabel(recommendation.reason_code)}
            </Button>
          </Card>
        ))}
      </div>
    </section>
  );
};
