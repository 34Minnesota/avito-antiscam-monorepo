export type Role = 'buyer' | 'seller';
export type Verdict = 'safe' | 'risky' | 'fatal';
export type Outcome = 'safe' | 'partial' | 'scammed';

export interface SessionResponse {
  sessionId: string;
}
export interface User {
  id: string;
  nickname: string;
  email: string;
  created_at: string;
}
export interface ScenarioStats {
  best_score: number;
  attempts_count: number;
}
export interface ScenarioCard {
  id: string;
  slug: string;
  role: Role;
  category: string;
  difficulty: number;
  title: string;
  description: string;
  stats?: ScenarioStats | null;
}
export interface Message {
  author: 'user' | 'counterpart' | 'system';
  text: string;
  attachment?: { type: string; title: string } | null;
}
export interface Listing {
  title: string;
  price: number;
  location: string;
  image: string;
}
export interface Counterpart {
  name: string;
  rating: number;
  reviews: number;
  registered: string;
}
export interface Option {
  id: string;
  text: string;
}
export interface Scene {
  scene_id: string;
  intro: Message[];
  prompt: string;
  options: Option[];
}
export interface StartResult {
  attempt_id: string;
  listing: Listing;
  counterpart: Counterpart;
  role: Role;
  title: string;
  scene: Scene;
  scenes_total: number;
}
export interface ChoiceResult {
  feedback: { verdict: Verdict; text: string };
  reaction: Message[];
  next_scene?: Scene | null;
  finished: boolean;
  summary?: Summary | null;
  revision: number;
}
export interface FlagInfo {
  id: string;
  title: string;
  text: string;
}
export interface Ending {
  outcome: Outcome;
  title: string;
  text: string;
}
export interface Summary {
  score: number;
  outcome: Outcome;
  ending: Ending;
  missed_flags: FlagInfo[];
  takeaway: string;
  steps_total: number;
  delta_vs_previous?: number | null;
}
export interface ScoreResponse {
  points: number;
  maxPoints: number;
  percent: number;
}
export interface CompletedAttempt {
  attemptId: string;
  score: ScoreResponse;
  outcome: Outcome;
  completedAt: string;
}
export interface ScenarioProgress {
  scenarioSlug: string;
  title: string;
  completed: boolean;
  passed: boolean;
  attemptsCount: number;
  bestScore?: ScoreResponse | null;
  activeAttemptId?: string | null;
  recentAttempts: CompletedAttempt[];
  initialScore?: ScoreResponse | null;
  latestScore?: ScoreResponse | null;
  improvementPercentPoints?: number | null;
  trend?: 'improving' | 'stable' | 'declining' | null;
  firstSafeAttempt?: CompletedAttempt | null;
}
export interface RoleProgress {
  role: Role;
  totalScenarios: number;
  completedScenarios: number;
  passedScenarios: number;
  completionPercent: number;
  passedPercent: number;
  scenarios: ScenarioProgress[];
}
export interface Progress {
  totalScenarios: number;
  completedScenarios: number;
  passedScenarios: number;
  completionPercent: number;
  passedPercent: number;
  roles: RoleProgress[];
  roleComparison: { completionPercentDelta: number; passedPercentDelta: number };
  recommendations: Array<{ scenarioSlug: string; reasonCode: string; reasonText: string }>;
}
