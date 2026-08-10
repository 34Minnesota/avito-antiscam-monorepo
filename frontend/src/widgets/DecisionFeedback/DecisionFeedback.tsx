import { useEffect, useRef } from 'react';
import { Verdict } from '@/shared/api/contracts';
import { Button } from '@/shared/ui/Button';
import cls from './DecisionFeedback.module.scss';

interface DecisionFeedbackProps {
  verdict: Verdict;
  text: string;
  onContinue: () => void;
  isFinal?: boolean;
}

const config: Record<Verdict, { title: string; icon: string }> = {
  safe: { title: 'Безопасное решение', icon: '✓' },
  risky: { title: 'Риск повышен', icon: '!' },
  fatal: { title: 'Опасное действие', icon: '×' },
};

export const DecisionFeedback = ({
  verdict,
  text,
  onContinue,
  isFinal = false,
}: DecisionFeedbackProps) => {
  const item = config[verdict];
  const buttonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    buttonRef.current?.focus();
  }, [verdict, text]);

  return (
    <section className={`${cls.card} ${cls[verdict]}`} aria-live="polite" aria-label={item.title}>
      <div className={cls.icon} aria-hidden="true">
        {item.icon}
      </div>
      <div className={cls.content}>
        <strong>{item.title}</strong>
        <p>{text}</p>
      </div>
      <Button ref={buttonRef} type="button" size="m" onClick={onContinue}>
        {isFinal ? 'Посмотреть итог' : 'Продолжить →'}
      </Button>
    </section>
  );
};
