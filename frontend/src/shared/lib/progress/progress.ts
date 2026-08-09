import { ScenarioProgress } from '@/shared/api/contracts';

type ProgressHeadlineInput = Pick<
  ScenarioProgress,
  'completed' | 'improvementPercentPoints' | 'latestScore' | 'activeAttemptId'
>;

export const getProgressHeadline = (progress: ProgressHeadlineInput) => {
  if (progress.activeAttemptId) return 'Тренировка не завершена';
  if (!progress.completed) return 'Ещё не пройден';
  if (progress.improvementPercentPoints && progress.improvementPercentPoints > 0)
    return `+${Math.round(progress.improvementPercentPoints)} баллов к первому результату`;
  if (progress.latestScore) return `${progress.latestScore.percent}% — последний результат`;
  return 'Пройден';
};
