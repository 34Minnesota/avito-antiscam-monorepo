import { useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { useGetMeQuery } from '@/entities/User';
import { useGetProgressQuery } from '@/entities/Progress';
import { useGetScenariosQuery } from '@/entities/Scenario';
import { isUnauthorized } from '@/shared/api/errors';
import { RoleSwitcher } from '@/features/roleSwitcher';
import { ProgressOverview } from '@/widgets/ProgressOverview/ProgressOverview';
import { ProgressInsights } from '@/widgets/ProgressInsights/ProgressInsights';
import { ProgressRecommendations } from '@/widgets/ProgressRecommendations/ProgressRecommendations';
import { ScenarioCatalog } from '@/widgets/ScenarioCatalog/ScenarioCatalog';
import { AppHeader } from '@/widgets/AppHeader/AppHeader';
import { Loader } from '@/shared/ui/Loader';
import { Role, ScenarioCard as ScenarioCardType } from '@/shared/api/contracts';
import { getStoredRole, saveStoredRole } from '@/shared/lib/role/storage';
import cls from './DashboardPage.module.scss';

const NavigateToLogin = () => {
  return <Navigate to="/login" replace />;
};

export const DashboardPage = () => {
  const navigate = useNavigate();
  const { data: user } = useGetMeQuery();
  const progressQuery = useGetProgressQuery();
  const {
    data: progress,
    isLoading: progressLoading,
    isFetching: progressFetching,
    error: progressError,
    refetch: refetchProgress,
  } = progressQuery;
  const [role, setRole] = useState<Role>(() => getStoredRole() ?? 'seller');
  const catalogQuery = useGetScenariosQuery(role);
  const {
    data: catalog,
    isLoading: catalogLoading,
    isFetching: catalogFetching,
    error: catalogError,
    refetch: refetchCatalog,
  } = catalogQuery;

  const changeRole = (nextRole: Role) => {
    setRole(nextRole);
    saveStoredRole(nextRole);
  };

  const dataLoading = progressLoading || catalogLoading || progressFetching || catalogFetching;
  const unauthorized = isUnauthorized(progressError) || isUnauthorized(catalogError);

  if (unauthorized) {
    return <NavigateToLogin />;
  }

  if (dataLoading || !user) {
    return (
      <>
        <AppHeader />
        <main className={cls.loading} aria-live="polite">
          <Loader />
          <span>Загружаем ваш прогресс…</span>
        </main>
      </>
    );
  }

  if (progressError || catalogError || !progress) {
    return (
      <>
        <AppHeader />
        <main className={cls.errorState} role="alert">
          <div className={cls.errorCard}>
            <div className={cls.kicker}>НЕ УДАЛОСЬ ЗАГРУЗИТЬ ДАННЫЕ</div>
            <h1>Прогресс временно недоступен</h1>
            <p>
              Сессия сохранена. Попробуйте обновить данные — приложение не будет бесконечно
              показывать загрузку.
            </p>
            <div className={cls.errorActions}>
              <button
                type="button"
                onClick={() => {
                  void refetchProgress();
                  void refetchCatalog();
                }}
              >
                Повторить
              </button>
              <button
                type="button"
                className={cls.secondaryButton}
                onClick={() => window.location.reload()}
              >
                Обновить страницу
              </button>
            </div>
          </div>
        </main>
      </>
    );
  }

  const roleProgress = progress.roles.find((item) => item.role === role);
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

        <ProgressInsights progress={progress} role={role} scenarios={scenarios} />

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

        <ProgressRecommendations
          role={role}
          recommendations={progress.recommendations}
          scenarios={scenarios}
          onStart={(scenario) => navigate(`/training/${scenario.id}`)}
        />

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
