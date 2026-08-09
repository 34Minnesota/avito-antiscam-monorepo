import { renderHook, act, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useTrainingSession } from './useTrainingSession';

const navigate = vi.fn();
const start = vi.fn();
const choose = vi.fn();

vi.mock('react-router-dom', () => ({ useNavigate: () => navigate }));
vi.mock('@/shared/auth/session', () => ({ getSessionId: () => 'session' }));
vi.mock('@/features/training/api', () => ({
  useStartAttemptMutation: () => [start],
  useSubmitChoiceMutation: () => [choose],
}));

const firstScene = {
  scene_id: 'scene-1',
  intro: [{ author: 'counterpart' as const, text: 'Начало' }],
  prompt: 'Решение?',
  options: [{ id: 'safe', text: 'Проверить' }],
};
const secondScene = {
  scene_id: 'scene-2',
  intro: [{ author: 'counterpart' as const, text: 'Следующий шаг' }],
  prompt: 'Дальше?',
  options: [{ id: 'done', text: 'Закончить' }],
};
const startResult = {
  attempt_id: 'attempt-1',
  listing: { title: 'Товар', price: 1000, location: 'Москва', image: '' },
  counterpart: { name: 'Алексей', rating: 4.9, reviews: 10, registered: '2024' },
  role: 'seller' as const,
  title: 'Сценарий',
  scene: firstScene,
  scenes_total: 2,
};

const resolved = (value: unknown) => ({ unwrap: () => Promise.resolve(value) });
const rejected = (value: unknown) => ({ unwrap: () => Promise.reject(value) });

beforeEach(() => {
  vi.clearAllMocks();
  sessionStorage.clear();
  localStorage.clear();
  start.mockReturnValue(resolved(startResult));
});

describe('useTrainingSession', () => {
  it('starts an attempt and exposes the server scene', async () => {
    const { result } = renderHook(() => useTrainingSession('scenario-1'));
    await waitFor(() => expect(result.current.training?.attempt_id).toBe('attempt-1'));
    expect(result.current.messages[0].text).toBe('Начало');
    expect(result.current.sceneNumber).toBe(1);
    expect(start).toHaveBeenCalledWith('scenario-1');
  });

  it('prevents a duplicate choice while a submission is active', async () => {
    let resolveChoice!: (value: unknown) => void;
    choose.mockReturnValue({
      unwrap: () =>
        new Promise((resolve) => {
          resolveChoice = resolve;
        }),
    });
    const { result } = renderHook(() => useTrainingSession('scenario-1'));
    await waitFor(() => expect(result.current.training).not.toBeNull());

    await act(async () => {
      void result.current.chooseOption('safe', 'Проверить');
    });
    expect(result.current.status).toBe('submitting');
    await act(async () => {
      void result.current.chooseOption('safe', 'Проверить ещё раз');
    });
    expect(choose).toHaveBeenCalledTimes(1);

    await act(async () => {
      resolveChoice({
        feedback: { verdict: 'safe', text: 'Ок' },
        reaction: [],
        next_scene: secondScene,
        finished: false,
        revision: 1,
      });
    });
  });

  it('recovers from a revision conflict by reloading the active attempt', async () => {
    choose.mockReturnValueOnce(rejected({ status: 409 }));
    const { result } = renderHook(() => useTrainingSession('scenario-1'));
    await waitFor(() => expect(result.current.training).not.toBeNull());

    await act(async () => {
      await result.current.chooseOption('safe', 'Проверить');
    });
    await waitFor(() => expect(start).toHaveBeenCalledTimes(2));
    expect(result.current.status).toBe('active');
    expect(result.current.errorMessage).toBeNull();
  });
});
