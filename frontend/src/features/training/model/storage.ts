import { Message } from '@/shared/api/contracts';
import { parseRevision } from '@/shared/lib/attempt/revision';

const ATTEMPT_PREFIX = 'antiscam:attempt:';
const SCENARIO_PREFIX = 'antiscam:attempt-scenario:';

export type AttemptSnapshot = {
  sceneNumber: number;
  sceneId?: string | null;
  messages: Message[];
  revision: number;
};

const key = (attemptId: string) => `${ATTEMPT_PREFIX}${attemptId}`;
const scenarioKey = (attemptId: string) => `${SCENARIO_PREFIX}${attemptId}`;

export const getAttemptSnapshot = (attemptId: string): AttemptSnapshot | null => {
  try {
    const raw = sessionStorage.getItem(key(attemptId));
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Partial<AttemptSnapshot>;
    if (!Array.isArray(parsed.messages)) return null;
    return {
      sceneNumber:
        Number.isFinite(parsed.sceneNumber) && Number(parsed.sceneNumber) > 0
          ? Number(parsed.sceneNumber)
          : 1,
      sceneId: typeof parsed.sceneId === 'string' ? parsed.sceneId : null,
      messages: parsed.messages,
      revision: parseRevision(parsed.revision),
    };
  } catch {
    return null;
  }
};

export const saveAttemptSnapshot = (attemptId: string, snapshot: AttemptSnapshot) => {
  try {
    sessionStorage.setItem(key(attemptId), JSON.stringify(snapshot));
  } catch {
    // Storage is a convenience cache; server state remains authoritative.
  }
};

export const clearAttemptSnapshot = (attemptId: string) => {
  sessionStorage.removeItem(key(attemptId));
  sessionStorage.removeItem(scenarioKey(attemptId));
};

export const saveAttemptScenario = (attemptId: string, scenarioId: string) => {
  try {
    sessionStorage.setItem(scenarioKey(attemptId), scenarioId);
  } catch {
    // Storage is a convenience cache; the server remains authoritative.
  }
};

export const getAttemptScenario = (attemptId: string): string | null => {
  try {
    return sessionStorage.getItem(scenarioKey(attemptId));
  } catch {
    return null;
  }
};
