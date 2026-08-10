import { Option } from '@/shared/api/contracts';
import cls from './TrainingDecision.module.scss';

interface Props {
  prompt: string;
  options: Option[];
  onChoose: (id: string) => void;
  disabled?: boolean;
}

export const TrainingDecision = ({ prompt, options, onChoose, disabled }: Props) => (
  <section className={cls.Decision} aria-labelledby="decision-title">
    <div className={cls.prompt} id="decision-title">
      {prompt}
    </div>
    <div className={cls.options} role="group" aria-label="Варианты решения">
      {options.map((option, index) => (
        <button
          type="button"
          disabled={disabled}
          key={option.id}
          onClick={() => onChoose(option.id)}
        >
          <span aria-hidden="true">{String.fromCharCode(65 + index)}</span>
          {option.text}
        </button>
      ))}
    </div>
    <p className={cls.hint} aria-live="polite">
      {disabled ? 'Сохраняем решение…' : 'Выберите один вариант'}
    </p>
  </section>
);
