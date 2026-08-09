import { ScenarioProgress } from '@/shared/api/contracts';

type ProgressHeadlineInput = Pick<
  ScenarioProgress,
  'completed' | 'improvement_percent_points' | 'latest_score' | 'active_attempt_id'
>;

export const getProgressHeadline = (progress: ProgressHeadlineInput) => {
  if (progress.active_attempt_id) return 'Тренировка не завершена';
  if (!progress.completed) return 'Ещё не пройден';
  if (progress.improvement_percent_points && progress.improvement_percent_points > 0)
    return `+${Math.round(progress.improvement_percent_points)} баллов к первому результату`;
  if (progress.latest_score) return `${progress.latest_score.percent}% — последний результат`;
  return 'Пройден';
};
