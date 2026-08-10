import { Progress, Role, ScenarioCard, ScenarioProgress } from '@/shared/api/contracts';

export interface ExperienceSummary {
  xp: number;
  level: number;
  currentLevelXp: number;
  nextLevelXp: number;
  progressPercent: number;
  completedScenarios: number;
  safeScenarios: number;
}

export interface SkillInsight {
  name: string;
  score: number;
  completed: number;
}

export interface ProgressDynamics {
  initial: number;
  latest: number;
  delta: number;
  trend: 'improving' | 'stable' | 'declining' | 'none';
  scenariosTracked: number;
}

const clamp = (value: number, min = 0, max = 100) => Math.min(max, Math.max(min, value));

const latestPercent = (scenario: ScenarioProgress) => {
  const value = scenario.latest_score?.percent ?? scenario.best_score?.percent ?? null;
  return typeof value === 'number' && Number.isFinite(value) ? clamp(value) : null;
};

const initialPercent = (scenario: ScenarioProgress) => {
  const explicit = scenario.initial_score?.percent;
  if (typeof explicit === 'number' && Number.isFinite(explicit)) return clamp(explicit);
  const attempts = [...scenario.recent_attempts].sort(
    (left, right) => Date.parse(left.completed_at) - Date.parse(right.completed_at),
  );
  const value = attempts[0]?.score?.percent;
  return typeof value === 'number' && Number.isFinite(value) ? clamp(value) : null;
};

const latestOrHistoryPercent = (scenario: ScenarioProgress) => {
  const explicit = latestPercent(scenario);
  if (explicit != null) return explicit;
  const attempts = [...scenario.recent_attempts].sort(
    (left, right) => Date.parse(left.completed_at) - Date.parse(right.completed_at),
  );
  const value = attempts.at(-1)?.score?.percent;
  return typeof value === 'number' && Number.isFinite(value) ? clamp(value) : null;
};

const SKILL_LABELS: Record<string, string> = {
  phishing_link: 'Проверка ссылок',
  qr_payment: 'Безопасная оплата',
  messenger_move: 'Общение в чате',
  sms_code: 'Защита кодов',
  fake_delivery: 'Проверка доставки',
  rent: 'Безопасная аренда',
  hacked_account: 'Защита аккаунта',
  overpayment: 'Проверка возврата',
  fee_scam: 'Проверка комиссий',
  prepayment: 'Защита от предоплаты',
  money_mule: 'Безопасность платежей',
};

export const getExperience = (progress: Progress, role: Role): ExperienceSummary => {
  const roleProgress = progress.roles.find((item) => item.role === role);
  const scenarios = roleProgress?.scenarios ?? [];
  const completed = scenarios.filter((scenario) => scenario.completed);
  const safe = completed.filter((scenario) => scenario.passed);
  const experience = progress.experience;

  const nextLevelXp = Math.max(1, experience.next_level_xp);
  const currentLevelXp = Math.min(nextLevelXp, Math.max(0, experience.current_xp));

  return {
    xp: Math.max(0, experience.total_xp),
    level: Math.max(1, experience.level),
    currentLevelXp,
    nextLevelXp,
    progressPercent: clamp((currentLevelXp / nextLevelXp) * 100),
    completedScenarios: completed.length,
    safeScenarios: safe.length,
  };
};

export const getSkills = (
  progress: Progress,
  role: Role,
  scenarios: ScenarioCard[],
): SkillInsight[] => {
  const roleProgress = progress.roles.find((item) => item.role === role);
  const bySlug = new Map(scenarios.map((scenario) => [scenario.slug, scenario]));
  const groups = new Map<string, number[]>();

  for (const item of roleProgress?.scenarios ?? []) {
    const score = latestPercent(item);
    const category = bySlug.get(item.scenario_slug)?.category;
    if (!category || score == null) continue;
    const values = groups.get(category) ?? [];
    values.push(score);
    groups.set(category, values);
  }

  return [...groups.entries()]
    .map(([category, scores]) => ({
      name: SKILL_LABELS[category] ?? category.replaceAll('_', ' '),
      score: Math.round(scores.reduce((sum, value) => sum + value, 0) / scores.length),
      completed: scores.length,
    }))
    .sort((a, b) => b.score - a.score);
};

export const getProgressDynamics = (progress: Progress, role: Role): ProgressDynamics => {
  const roleProgress = progress.roles.find((item) => item.role === role);
  const tracked = (roleProgress?.scenarios ?? [])
    .map((scenario) => ({
      scenario,
      initial: initialPercent(scenario),
      latest: latestOrHistoryPercent(scenario),
    }))
    .filter((item) => item.initial != null && item.latest != null);

  if (!tracked.length) {
    return { initial: 0, latest: 0, delta: 0, trend: 'none', scenariosTracked: 0 };
  }

  const initial = Math.round(
    tracked.reduce((sum, item) => sum + (item.initial ?? 0), 0) / tracked.length,
  );
  const latest = Math.round(
    tracked.reduce((sum, item) => sum + (item.latest ?? 0), 0) / tracked.length,
  );
  const delta = latest - initial;
  const trend = delta > 2 ? 'improving' : delta < -2 ? 'declining' : 'stable';

  return { initial, latest, delta, trend, scenariosTracked: tracked.length };
};

export const getTrendLabel = (trend: ProgressDynamics['trend']) => {
  if (trend === 'improving') return 'Результат растёт';
  if (trend === 'declining') return 'Нужна дополнительная практика';
  if (trend === 'stable') return 'Результат стабилен';
  return 'Пока нет динамики';
};

export const getCategoryLabel = (category: string) =>
  SKILL_LABELS[category] ?? category.replaceAll('_', ' ');
