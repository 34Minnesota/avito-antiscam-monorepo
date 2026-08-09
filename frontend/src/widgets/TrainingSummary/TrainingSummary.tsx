import { Summary } from '@/shared/api/contracts';
import { Badge } from '@/shared/ui/Badge';
import { Button } from '@/shared/ui/Button';
import { Card } from '@/shared/ui/Card';
import cls from './TrainingSummary.module.scss';

const outcome = (value: Summary['outcome']) => {
  if (value === 'safe') return { label: 'Безопасно', tone: 'good' as const };
  if (value === 'partial') return { label: 'Есть риск', tone: 'warn' as const };
  return { label: 'Мошенничество', tone: 'danger' as const };
};

interface TrainingSummaryProps {
  summary: Summary;
  onRetry: () => void;
  onDashboard: () => void;
}

export const TrainingSummary = ({ summary, onRetry, onDashboard }: TrainingSummaryProps) => {
  const result = outcome(summary.outcome);
  const missedFlags = Array.isArray(summary.missed_flags) ? summary.missed_flags : [];
  const score = Math.max(0, Math.min(100, Number(summary.score) || 0));
  const steps = Number(summary.steps_total) || 0;

  return (
    <div className={cls.page}>
      <div className={cls.score}>
        <div
          className={cls.ring}
          style={{
            background: `radial-gradient(circle,var(--summary-ring-bg) 58%,transparent 59%), conic-gradient(var(--accent-orange) ${score}%,var(--summary-ring-track) 0)`,
          }}
        >
          <strong>{score}</strong>
          <span>/100</span>
        </div>
        <Badge tone={result.tone}>{result.label}</Badge>
        <h1>{summary.ending?.title ?? 'Сценарий завершён'}</h1>
        <p>{summary.ending?.text ?? 'Разбор попытки готов.'}</p>
        {summary.delta_vs_previous != null && (
          <div className={summary.delta_vs_previous >= 0 ? cls.deltaGood : cls.deltaBad}>
            {summary.delta_vs_previous >= 0 ? '+' : ''}
            {summary.delta_vs_previous} баллов к прошлому результату
          </div>
        )}
      </div>

      <div className={cls.columns}>
        <Card className={cls.card}>
          <div className={cls.label}>ЧТО ВЫ ПРОПУСТИЛИ</div>
          {missedFlags.length ? (
            missedFlags.map((flag) => (
              <div className={cls.flag} key={flag.id}>
                <span>!</span>
                <div>
                  <b>{flag.title}</b>
                  <p>{flag.text}</p>
                </div>
              </div>
            ))
          ) : (
            <div className={cls.safe}>✓ Вы не пропустили ключевые признаки мошенничества</div>
          )}
        </Card>

        <Card className={cls.card}>
          <div className={cls.label}>ГЛАВНОЕ ПРАВИЛО</div>
          <p className={cls.takeaway}>
            {summary.takeaway || 'Проверяйте факты и не принимайте решения под давлением.'}
          </p>
          <div className={cls.steps}>
            Пройдено решений: <b>{steps}</b>
          </div>
        </Card>
      </div>

      <div className={cls.actions}>
        <Button type="button" size="l" onClick={onRetry}>
          Пройти ещё раз
        </Button>
        <Button type="button" size="l" variant="secondary" onClick={onDashboard}>
          На главную
        </Button>
      </div>
    </div>
  );
};
