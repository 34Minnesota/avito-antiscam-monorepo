import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useGetMeQuery } from '@/entities/User';
import { useGetProgressQuery } from '@/entities/Progress';
import { useGetScenariosQuery } from '@/entities/Scenario';
import { RoleSwitcher } from '@/features/roleSwitcher';
import { ProgressOverview } from '@/widgets/ProgressOverview/ProgressOverview';
import { ScenarioCatalog } from '@/widgets/ScenarioCatalog/ScenarioCatalog';
import { AppHeader } from '@/widgets/AppHeader/AppHeader';
import { Loader } from '@/shared/ui/Loader';
import { Role, ScenarioCard as ScenarioCardType } from '@/shared/api/contracts';
import { getStoredRole, saveStoredRole } from '@/shared/lib/role/storage';
import cls from './DashboardPage.module.scss';

export const DashboardPage = () => {
  const navigate = useNavigate();
  const { data: user } = useGetMeQuery();
  const { data: progress, isLoading: progressLoading } = useGetProgressQuery();
  const [role, setRole] = useState<Role>(() => getStoredRole() ?? 'seller');
  const { data: catalog, isLoading: catalogLoading } = useGetScenariosQuery(role);

  const changeRole = (nextRole: Role) => {
    setRole(nextRole);
    saveStoredRole(nextRole);
  };

  if (progressLoading || catalogLoading || !user || !progress) {
    return (
      <>
        <AppHeader />
        <main className={cls.loading}>
          <Loader />
        </main>
      </>
    );
  }

  const roleProgress = progress.roles.find((item) => item.role === role);
  const recommendation =
    progress.recommendations.find((item) =>
      roleProgress?.scenarios.some((scenario) => scenario.scenarioSlug === item.scenarioSlug),
    ) ?? progress.recommendations[0];
  const recommendationScenario = catalog?.scenarios.find(
    (item) => item.slug === recommendation?.scenarioSlug,
  );
  const scenarios = catalog?.scenarios ?? [];

  return (
    <>
      <AppHeader />
      <main className={cls.page}>
        <section className={cls.hero}>
          <div>
            <div className={cls.kicker}>ТРЕНАЖЁР БЕЗОПАСНЫХ СДЕЛОК</div>
            <h1>Привет, {user.nickname}.</h1>
            <p>
              Тренируйте решения, которые помогают распознавать мошенничество до того, как оно
              становится проблемой.
            </p>
          </div>
          <RoleSwitcher value={role} onChange={changeRole} />
        </section>

        <ProgressOverview progress={progress} activeRole={role} />

        <section className={cls.learning}>
          <div className={cls.learningIcon}>✓</div>
          <div>
            <strong>Как проходит тренировка</strong>
            <p>
              Читайте диалог, принимайте решение в спорной точке и сразу видите последствия. После
              завершения разберите пропущенные признаки и попробуйте ещё раз.
            </p>
          </div>
          <div className={cls.learningSteps}>
            <span>01 Диалог</span>
            <span>02 Решение</span>
            <span>03 Разбор</span>
          </div>
        </section>

        {recommendation && recommendationScenario && (
          <section className={cls.recommendation} aria-label="Рекомендованная тренировка">
            <div>
              <span>СЛЕДУЮЩИЙ ШАГ</span>
              <h3>{recommendation.reasonText}</h3>
              <p>
                Рекомендация основана на вашем прогрессе и помогает закрыть следующий пробел в
                навыках.
              </p>
            </div>
            <button
              type="button"
              onClick={() => navigate(`/training/${recommendationScenario.id}`)}
            >
              Начать тренировку →
            </button>
          </section>
        )}

        <section className={cls.catalog}>
          <div className={cls.sectionHead}>
            <div>
              <div className={cls.sectionTitle}>
                Сценарии {role === 'seller' ? 'продавца' : 'покупателя'}
              </div>
              <p>Ситуации похожи на реальные сделки и не сводятся к очевидным вопросам.</p>
            </div>
            <span>
              {scenarios.length} {scenarios.length === 1 ? 'сценарий' : 'сценариев'}
            </span>
          </div>
          <ScenarioCatalog
            scenarios={scenarios}
            progress={roleProgress?.scenarios}
            onStart={(scenario: ScenarioCardType) => {
              navigate(`/training/${scenario.id}`);
            }}
          />
        </section>
      </main>
    </>
  );
};
