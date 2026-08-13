export type Role = 'buyer' | 'seller';
export type Verdict = 'safe' | 'risky' | 'fatal';
export type Outcome = 'safe' | 'partial' | 'scammed';

export interface SessionResponse {
  session_id: string;
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
  id?: string;
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
  revision?: number;
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
  max_points: number;
  percent: number;
}
export interface CompletedAttempt {
  attempt_id: string;
  score: ScoreResponse;
  outcome: Outcome;
  completed_at: string;
}
export interface ScenarioProgress {
  scenario_slug: string;
  title: string;
  completed: boolean;
  passed: boolean;
  attempts_count: number;
  best_score?: ScoreResponse | null;
  active_attempt_id?: string | null;
  recent_attempts: CompletedAttempt[];
  initial_score?: ScoreResponse | null;
  latest_score?: ScoreResponse | null;
  improvement_percent_points?: number | null;
  trend?: 'improving' | 'stable' | 'declining' | null;
  first_safe_attempt?: CompletedAttempt | null;
}
export interface RoleProgress {
  role: Role;
  total_scenarios: number;
  completed_scenarios: number;
  passed_scenarios: number;
  completion_percent: number;
  passed_percent: number;
  scenarios: ScenarioProgress[];
}
export interface Achievement {
  code: string;
  title: string;
  description: string;
  earned: boolean;
}

export interface ExperienceProgress {
  total_xp: number;
  level: number;
  current_xp: number;
  next_level_xp: number;
  achievements: Achievement[];
}

export interface ProgressRecommendation {
  role: Role;
  scenario_slug: string;
  reason_code: string;
  reason_text: string;
}

export interface Progress {
  total_scenarios: number;
  completed_scenarios: number;
  passed_scenarios: number;
  completion_percent: number;
  passed_percent: number;
  roles: RoleProgress[];
  role_comparison: { completion_percent_delta: number; passed_percent_delta: number };
  recommendations: ProgressRecommendation[];
  experience: ExperienceProgress;
}
