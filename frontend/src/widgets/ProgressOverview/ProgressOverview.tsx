import { Progress, Role } from '@/shared/api/contracts';
import { getRoleComparison } from '@/shared/lib/progress/comparison';
import { Card } from '@/shared/ui/Card';
import cls from './ProgressOverview.module.scss';

interface ProgressOverviewProps {
  progress: Progress;
  activeRole: Role;
}

const formatAttemptDate = (value: string) => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'Дата неизвестна';
  return new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'short' }).format(date);
};

export const ProgressOverview = ({ progress, activeRole }: ProgressOverviewProps) => {
  const role = progress.roles.find((item) => item.role === activeRole);
  const comparisonContent = getRoleComparison(progress.roles);
  const recentAttempts = (role?.scenarios ?? [])
    .flatMap((scenario) =>
      scenario.recent_attempts.map((attempt) => ({ ...attempt, title: scenario.title })),
    )
    .sort((left, right) => Date.parse(right.completed_at) - Date.parse(left.completed_at))
    .slice(0, 4);

  return (
    <div className={cls.grid}>
      <Card className={cls.hero}>
        <div className={cls.eyebrow}>
          ПРОГРЕСС · {activeRole === 'seller' ? 'ПРОДАВЕЦ' : 'ПОКУПАТЕЛЬ'}
        </div>
        <div className={cls.percent}>
          {role?.completion_percent ?? 0}
          <small>%</small>
        </div>
        <div
          className={cls.track}
          role="progressbar"
          aria-valuenow={role?.completion_percent ?? 0}
          aria-valuemin={0}
          aria-valuemax={100}
        >
          <span style={{ width: `${role?.completion_percent ?? 0}%` }} />
        </div>
        <p>
          {role?.completed_scenarios ?? 0} из {role?.total_scenarios ?? 0} сценариев пройдено
        </p>
      </Card>
      <Card className={cls.stat}>
        <span>Безопасные решения</span>
        <strong>{role?.passed_percent ?? 0}%</strong>
        <small>{role?.passed_scenarios ?? 0} успешных прохождений</small>
      </Card>
      <Card className={cls.stat}>
        <span>Сравнение прогресса</span>
        <strong>{comparisonContent.value}</strong>
        <small>{comparisonContent.note}</small>
      </Card>

      <Card className={cls.history}>
        <div className={cls.historyHead}>
          <div>
            <span className={cls.historyEyebrow}>ПОСЛЕДНИЕ РЕЗУЛЬТАТЫ</span>
            <h3>Ваш прогресс в динамике</h3>
          </div>
          {recentAttempts.length > 0 && (
            <span className={cls.historyCount}>{recentAttempts.length} последних</span>
          )}
        </div>
        {recentAttempts.length > 0 ? (
          <div className={cls.historyList} aria-label="Последние результаты">
            {recentAttempts.map((attempt) => (
              <div className={cls.historyItem} key={attempt.attempt_id}>
                <div>
                  <strong>{attempt.score.percent}%</strong>
                  <span>{attempt.title}</span>
                </div>
                <div className={cls.historyMeta}>
                  <span>
                    {attempt.outcome === 'safe'
                      ? 'Безопасно'
                      : attempt.outcome === 'partial'
                        ? 'Частично'
                        : 'Рискованно'}
                  </span>
                    <small>{formatAttemptDate(attempt.completed_at)}</small>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className={cls.historyEmpty}>
            Завершите первую тренировку — здесь появится история ваших результатов.
          </p>
        )}
      </Card>
    </div>
  );
};
