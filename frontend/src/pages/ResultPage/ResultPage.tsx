import { useNavigate, useParams } from 'react-router-dom';
import { useGetSummaryQuery } from '@/features/training';
import { getAttemptScenario } from '@/features/training/model/storage';
import { TrainingSummary } from '@/widgets/TrainingSummary/TrainingSummary';
import { AppHeader } from '@/widgets/AppHeader/AppHeader';
import { Loader } from '@/shared/ui/Loader';
import { Button } from '@/shared/ui/Button';
import cls from './ResultPage.module.scss';

export const ResultPage = () => {
  const { attemptId } = useParams();
  const navigate = useNavigate();
  const {
    data: summary,
    isLoading,
    error,
  } = useGetSummaryQuery(attemptId ?? '', { skip: !attemptId });
  const scenarioId = attemptId ? getAttemptScenario(attemptId) : null;

  if (isLoading)
    return (
      <>
        <AppHeader />
        <main className={cls.empty}>
          <Loader />
        </main>
      </>
    );

  if (error || !summary || !summary.ending) {
    return (
      <>
        <AppHeader />
        <main className={cls.empty}>
          <h2>Результат не найден</h2>
          <p>Попробуйте открыть сценарий из каталога ещё раз.</p>
          <Button type="button" onClick={() => navigate('/')}>
            Вернуться к сценариям
          </Button>
        </main>
      </>
    );
  }

  return (
    <>
      <AppHeader />
      <main className={cls.page}>
        <TrainingSummary
          summary={summary}
          onRetry={() => {
            if (scenarioId) {
              navigate(`/training/${scenarioId}`);
            } else {
              navigate('/');
            }
          }}
          onDashboard={() => navigate('/')}
        />
      </main>
    </>
  );
};
