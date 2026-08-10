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
  const completion = Math.min(100, Math.max(0, role?.completion_percent ?? 0));
  const passed = Math.min(100, Math.max(0, role?.passed_percent ?? 0));
  const comparisonContent = getRoleComparison(progress.roles);
  const recentAttempts = (role?.scenarios ?? [])
    .flatMap((scenario) =>
      scenario.recent_attempts.map((attempt) => ({
        ...attempt,
        title: scenario.title,
        firstSafe: scenario.first_safe_attempt?.attempt_id === attempt.attempt_id,
      })),
    )
    .sort((left, right) => Date.parse(right.completed_at) - Date.parse(left.completed_at))
    .slice(0, 4);
  const firstSafe = (role?.scenarios ?? [])
    .map((scenario) => scenario.first_safe_attempt)
    .filter((attempt): attempt is NonNullable<typeof attempt> => Boolean(attempt))
    .sort((left, right) => Date.parse(left.completed_at) - Date.parse(right.completed_at))[0];

  return (
    <div className={cls.grid}>
      <Card className={cls.hero}>
        <div className={cls.eyebrow}>
          ПРОГРЕСС · {activeRole === 'seller' ? 'ПРОДАВЕЦ' : 'ПОКУПАТЕЛЬ'}
        </div>
        <div className={cls.percent}>
          {completion}
          <small>%</small>
        </div>
        <div
          className={cls.track}
          role="progressbar"
          aria-valuenow={completion}
          aria-valuemin={0}
          aria-valuemax={100}
        >
          <span style={{ width: `${completion}%` }} />
        </div>
        <p>
          {role?.completed_scenarios ?? 0} из {role?.total_scenarios ?? 0} сценариев пройдено
        </p>
      </Card>
      <Card className={cls.stat}>
        <span>Безопасные прохождения</span>
        <strong>{passed}%</strong>
        <small>{role?.passed_scenarios ?? 0} безопасных прохождений</small>
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
                    {attempt.firstSafe
                      ? 'Первый безопасный'
                      : attempt.outcome === 'safe'
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
        {firstSafe && (
          <div className={cls.firstSafe}>
            <span>✓</span>
            <div>
              <strong>Первый безопасный результат</strong>
              <small>
                {firstSafe.score.percent}% · {formatAttemptDate(firstSafe.completed_at)}
              </small>
            </div>
          </div>
        )}
      </Card>
    </div>
  );
};
