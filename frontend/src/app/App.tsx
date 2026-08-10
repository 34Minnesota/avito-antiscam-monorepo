import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { useGetMeQuery } from '@/entities/User';
import { clearSessionId, getSessionId } from '@/shared/auth/session';
import { isUnauthorized } from '@/shared/api/errors';
import { DashboardPage } from '@/pages/DashboardPage/DashboardPage';
import { AuthPage } from '@/pages/AuthPage/AuthPage';
import { TrainingPage } from '@/pages/TrainingPage/TrainingPage';
import { ResultPage } from '@/pages/ResultPage/ResultPage';
import { NotFoundPage } from '@/pages/NotFoundPage/NotFoundPage';
import { Loader } from '@/shared/ui/Loader';
import cls from './App.module.scss';

interface ProtectedProps {
  children: JSX.Element;
}

const Protected = ({ children }: ProtectedProps) => {
  const hasSession = Boolean(getSessionId());
  const { isLoading, isFetching, isSuccess, error, refetch } = useGetMeQuery(undefined, {
    skip: !hasSession,
  });
  const unauthorized = isUnauthorized(error);

  if (!hasSession || unauthorized) {
    if (unauthorized) clearSessionId();
    return <Navigate to="/login" replace />;
  }

  if (isLoading || isFetching) {
    return (
      <div className={cls.loading}>
        <Loader />
      </div>
    );
  }

  if (error || !isSuccess) {
    return (
      <div className={cls.loading}>
        <div>Не удалось проверить сессию.</div>
        <button type="button" onClick={() => void refetch()}>
          Повторить
        </button>
      </div>
    );
  }

  return children;
};

export const App = () => (
  <BrowserRouter>
    <Routes>
      <Route path="/login" element={<AuthPage />} />
      <Route
        path="/"
        element={
          <Protected>
            <DashboardPage />
          </Protected>
        }
      />
      <Route
        path="/training/:scenarioId"
        element={
          <Protected>
            <TrainingPage />
          </Protected>
        }
      />
      <Route
        path="/result/:attemptId"
        element={
          <Protected>
            <ResultPage />
          </Protected>
        }
      />
      <Route path="/404" element={<NotFoundPage />} />
      <Route path="*" element={<Navigate to="/404" replace />} />
    </Routes>
  </BrowserRouter>
);
