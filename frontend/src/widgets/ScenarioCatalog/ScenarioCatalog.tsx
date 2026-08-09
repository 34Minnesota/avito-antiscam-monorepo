import { ScenarioCard as ScenarioCardType, ScenarioProgress } from '@/shared/api/contracts';
import { ScenarioCard } from '@/entities/Scenario/ScenarioCard';
import cls from './ScenarioCatalog.module.scss';

interface Props {
  scenarios: ScenarioCardType[];
  progress?: ScenarioProgress[];
  onStart: (scenario: ScenarioCardType) => void;
}

export const ScenarioCatalog = ({ scenarios, progress = [], onStart }: Props) => (
  <div className={cls.grid}>
    {scenarios.map((scenario) => (
      <ScenarioCard
        key={scenario.id}
        scenario={scenario}
        progress={progress.find((item) => item.scenarioSlug === scenario.slug)}
        onStart={() => onStart(scenario)}
      />
    ))}
  </div>
);
