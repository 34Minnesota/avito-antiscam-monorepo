import { beforeEach, describe, expect, it } from 'vitest';
import {
  clearAttemptSnapshot,
  getAttemptScenario,
  getAttemptSnapshot,
  saveAttemptScenario,
  saveAttemptSnapshot,
} from './storage';

beforeEach(() => sessionStorage.clear());

describe('training storage', () => {
  it('round-trips an attempt snapshot', () => {
    saveAttemptSnapshot('a1', {
      sceneNumber: 2,
      sceneId: 'scene-2',
      messages: [{ author: 'counterpart', text: 'Привет' }],
      revision: 2,
    });
    expect(getAttemptSnapshot('a1')).toEqual({
      sceneNumber: 2,
      sceneId: 'scene-2',
      messages: [{ author: 'counterpart', text: 'Привет' }],
      revision: 2,
    });
    clearAttemptSnapshot('a1');
    expect(getAttemptSnapshot('a1')).toBeNull();
  });

  it('stores scenario ownership for a completed attempt', () => {
    saveAttemptScenario('a1', 'scenario-1');
    expect(getAttemptScenario('a1')).toBe('scenario-1');
  });

  it('clears scenario ownership together with the attempt snapshot', () => {
    saveAttemptScenario('a1', 'scenario-1');
    saveAttemptSnapshot('a1', { sceneNumber: 1, sceneId: 'scene-1', messages: [], revision: 0 });
    clearAttemptSnapshot('a1');
    expect(getAttemptScenario('a1')).toBeNull();
  });
});
