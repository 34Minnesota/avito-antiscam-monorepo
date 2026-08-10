import type {
  ChoiceResult,
  Progress,
  Role,
  RoleProgress,
  ScenarioCard,
  ScenarioProgress,
  StartResult,
  Summary,
  User,
} from './contracts';

const asRecord = (value: unknown, name: string): Record<string, unknown> => {
  if (typeof value !== 'object' || value === null) throw new Error(`Invalid API response: ${name}`);
  return value as Record<string, unknown>;
};

const isString = (value: unknown): value is string => typeof value === 'string';
const isNumber = (value: unknown): value is number =>
  typeof value === 'number' && Number.isFinite(value);

const numberFrom = (
  record: Record<string, unknown>,
  snake: string,
  camel = snake,
): number | undefined => {
  const value = record[snake] ?? record[camel];
  return isNumber(value) ? value : undefined;
};

const intFrom = (record: Record<string, unknown>, snake: string, camel = snake, fallback = 0) =>
  Math.max(0, Math.round(numberFrom(record, snake, camel) ?? fallback));

const percentFrom = (record: Record<string, unknown>, snake: string, camel = snake, fallback = 0) =>
  Math.min(100, Math.max(0, intFrom(record, snake, camel, fallback)));

const normalizeScore = (value: unknown) => {
  if (value == null) return value;
  const record = asRecord(value, 'Score');
  const points = intFrom(record, 'points');
  const maxPoints = Math.max(1, intFrom(record, 'max_points', 'maxPoints', 100));
  const percent = percentFrom(record, 'percent');
  return { points, max_points: maxPoints, percent };
};

const normalizeCompletedAttempt = (value: unknown) => {
  const record = asRecord(value, 'CompletedAttempt');
  if (!isString(record.attempt_id ?? record.attemptId))
    throw new Error('Invalid API response: CompletedAttempt');
  return {
    attempt_id: String(record.attempt_id ?? record.attemptId),
    score: normalizeScore(record.score)!,
    outcome: String(record.outcome) as 'safe' | 'partial' | 'scammed',
    completed_at: String(record.completed_at ?? record.completedAt ?? ''),
  };
};

const normalizeScenarioProgress = (value: unknown): ScenarioProgress => {
  const record = asRecord(value, 'ScenarioProgress');
  const slug = record.scenario_slug ?? record.scenarioSlug;
  if (!isString(slug)) throw new Error('Invalid API response: ScenarioProgress');

  const recentRaw = Array.isArray(record.recent_attempts)
    ? record.recent_attempts
    : Array.isArray(record.recentAttempts)
      ? record.recentAttempts
      : [];

  const normalizeOptionalAttempt = (attempt: unknown) =>
    attempt == null ? null : normalizeCompletedAttempt(attempt);

  return {
    scenario_slug: slug,
    title: isString(record.title) ? record.title : 'Сценарий',
    completed: Boolean(record.completed),
    passed: Boolean(record.passed),
    attempts_count: intFrom(record, 'attempts_count', 'attemptsCount'),
    best_score: normalizeScore(
      record.best_score ?? record.bestScore,
    ) as ScenarioProgress['best_score'],
    active_attempt_id: (record.active_attempt_id ?? record.activeAttemptId ?? null) as
      string | null,
    recent_attempts: recentRaw.map(normalizeCompletedAttempt),
    initial_score: normalizeScore(
      record.initial_score ?? record.initialScore,
    ) as ScenarioProgress['initial_score'],
    latest_score: normalizeScore(
      record.latest_score ?? record.latestScore,
    ) as ScenarioProgress['latest_score'],
    improvement_percent_points:
      numberFrom(record, 'improvement_percent_points', 'improvementPercentPoints') == null
        ? null
        : intFrom(record, 'improvement_percent_points', 'improvementPercentPoints'),
    trend: (record.trend ?? null) as ScenarioProgress['trend'],
    first_safe_attempt: normalizeOptionalAttempt(
      record.first_safe_attempt ?? record.firstSafeAttempt,
    ),
  };
};

const normalizeRoleProgress = (value: unknown): RoleProgress => {
  const record = asRecord(value, 'RoleProgress');
  const role = record.role;
  if (role !== 'buyer' && role !== 'seller')
    throw new Error('Invalid API response: RoleProgress.role');

  const rawScenarios = Array.isArray(record.scenarios) ? record.scenarios : [];
  const scenarios = rawScenarios.map(normalizeScenarioProgress);
  const totalScenarios = intFrom(record, 'total_scenarios', 'totalScenarios', scenarios.length);
  const completedScenarios = Math.min(
    totalScenarios,
    intFrom(
      record,
      'completed_scenarios',
      'completedScenarios',
      scenarios.filter((s) => s.completed).length,
    ),
  );
  const passedScenarios = Math.min(
    completedScenarios,
    intFrom(
      record,
      'passed_scenarios',
      'passedScenarios',
      scenarios.filter((s) => s.passed).length,
    ),
  );

  return {
    role: role as Role,
    total_scenarios: totalScenarios,
    completed_scenarios: completedScenarios,
    passed_scenarios: passedScenarios,
    completion_percent: percentFrom(
      record,
      'completion_percent',
      'completionPercent',
      totalScenarios ? Math.round((completedScenarios / totalScenarios) * 100) : 0,
    ),
    passed_percent: percentFrom(
      record,
      'passed_percent',
      'passedPercent',
      totalScenarios ? Math.round((passedScenarios / totalScenarios) * 100) : 0,
    ),
    scenarios,
  };
};

const normalizeProgress = (value: unknown): Progress => {
  const record = asRecord(value, 'Progress');
  if (!Array.isArray(record.roles)) throw new Error('Invalid API response: Progress.roles');

  const roles = record.roles.map(normalizeRoleProgress);
  const totalScenarios = intFrom(
    record,
    'total_scenarios',
    'totalScenarios',
    roles.reduce((sum, role) => sum + role.total_scenarios, 0),
  );
  const completedScenarios = intFrom(
    record,
    'completed_scenarios',
    'completedScenarios',
    roles.reduce((sum, role) => sum + role.completed_scenarios, 0),
  );
  const passedScenarios = intFrom(
    record,
    'passed_scenarios',
    'passedScenarios',
    roles.reduce((sum, role) => sum + role.passed_scenarios, 0),
  );

  const rawExperience = asRecord(record.experience ?? {}, 'Progress.experience');
  const rawComparison = asRecord(
    record.role_comparison ?? record.roleComparison ?? {},
    'Progress.roleComparison',
  );
  const rawRecommendations = Array.isArray(record.recommendations) ? record.recommendations : [];
  const rawAchievements = Array.isArray(rawExperience.achievements)
    ? rawExperience.achievements
    : [];

  return {
    total_scenarios: totalScenarios,
    completed_scenarios: completedScenarios,
    passed_scenarios: passedScenarios,
    completion_percent: percentFrom(
      record,
      'completion_percent',
      'completionPercent',
      totalScenarios ? Math.round((completedScenarios / totalScenarios) * 100) : 0,
    ),
    passed_percent: percentFrom(
      record,
      'passed_percent',
      'passedPercent',
      totalScenarios ? Math.round((passedScenarios / totalScenarios) * 100) : 0,
    ),
    roles,
    role_comparison: {
      completion_percent_delta: intFrom(
        rawComparison,
        'completion_percent_delta',
        'completionPercentDelta',
        (roles.find((r) => r.role === 'seller')?.completion_percent ?? 0) -
          (roles.find((r) => r.role === 'buyer')?.completion_percent ?? 0),
      ),
      passed_percent_delta: intFrom(
        rawComparison,
        'passed_percent_delta',
        'passedPercentDelta',
        (roles.find((r) => r.role === 'seller')?.passed_percent ?? 0) -
          (roles.find((r) => r.role === 'buyer')?.passed_percent ?? 0),
      ),
    },
    recommendations: rawRecommendations.map((item) => {
      const recommendation = asRecord(item, 'ProgressRecommendation');
      const role = recommendation.role;
      if (role !== 'buyer' && role !== 'seller')
        throw new Error('Invalid API response: ProgressRecommendation.role');

      return {
        role,
        scenario_slug: String(recommendation.scenario_slug ?? recommendation.scenarioSlug ?? ''),
        reason_code: String(recommendation.reason_code ?? recommendation.reasonCode ?? ''),
        reason_text: String(recommendation.reason_text ?? recommendation.reasonText ?? ''),
      };
    }),
    experience: {
      total_xp: intFrom(rawExperience, 'total_xp', 'totalXP'),
      level: Math.max(1, intFrom(rawExperience, 'level', 'level', 1)),
      current_xp: intFrom(rawExperience, 'current_xp', 'currentXP'),
      next_level_xp: Math.max(1, intFrom(rawExperience, 'next_level_xp', 'nextLevelXP', 100)),
      achievements: rawAchievements.map((item) => {
        const achievement = asRecord(item, 'Achievement');
        return {
          code: String(achievement.code ?? ''),
          title: String(achievement.title ?? ''),
          description: String(achievement.description ?? ''),
          earned: Boolean(achievement.earned),
        };
      }),
    },
  };
};

export const validateUser = (value: unknown): User => {
  const record = asRecord(value, 'User');
  if (!isString(record.id) || !isString(record.nickname) || !isString(record.email)) {
    throw new Error('Invalid API response: User');
  }
  return value as User;
};

export const validateScenarioList = (value: unknown): { scenarios: ScenarioCard[] } => {
  const record = asRecord(value, 'ScenarioList');
  if (!Array.isArray(record.scenarios)) throw new Error('Invalid API response: scenarios');
  return value as { scenarios: ScenarioCard[] };
};

export const validateProgress = (value: unknown): Progress => normalizeProgress(value);

export const validateStartResult = (value: unknown): StartResult => {
  const record = asRecord(value, 'StartResult');
  if (
    !isString(record.attempt_id) ||
    !isString(record.role) ||
    typeof record.scene !== 'object' ||
    record.scene === null
  ) {
    throw new Error('Invalid API response: StartResult');
  }
  return value as StartResult;
};

export const validateChoiceResult = (value: unknown): ChoiceResult => {
  const record = asRecord(value, 'ChoiceResult');
  if (
    typeof record.feedback !== 'object' ||
    record.feedback === null ||
    !Array.isArray(record.reaction) ||
    typeof record.finished !== 'boolean' ||
    !isNumber(record.revision)
  ) {
    throw new Error('Invalid API response: ChoiceResult');
  }
  return value as ChoiceResult;
};

export const validateSummary = (value: unknown): Summary => {
  const record = asRecord(value, 'Summary');
  if (
    !isNumber(record.score) ||
    !isString(record.outcome) ||
    typeof record.ending !== 'object' ||
    record.ending === null
  ) {
    throw new Error('Invalid API response: Summary');
  }
  return value as Summary;
};
