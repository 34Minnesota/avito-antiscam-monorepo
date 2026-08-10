import { renderHook, act, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useTrainingSession } from './useTrainingSession';

const navigate = vi.fn();
const start = vi.fn();
const choose = vi.fn();

vi.mock('react-router-dom', () => ({ useNavigate: () => navigate }));
vi.mock('@/shared/auth/session', () => ({
  getSessionId: () => 'session',
  clearSessionId: vi.fn(),
}));
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
  it('redirects to login when starting is unauthorized', async () => {
    start.mockReturnValueOnce(rejected({ status: 401 }));
    renderHook(() => useTrainingSession('scenario-1'));
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/login', { replace: true }));
  });

  it('shows a recoverable error when starting fails', async () => {
    start.mockReturnValueOnce(rejected({ status: 500 }));
    const { result } = renderHook(() => useTrainingSession('scenario-1'));
    await waitFor(() => expect(result.current.status).toBe('error'));
    expect(result.current.errorMessage).toContain('Не удалось открыть сценарий');
  });

  it('shows an error when the server omits the next scene', async () => {
    choose.mockReturnValueOnce(
      resolved({
        feedback: { verdict: 'risky', text: 'Проверьте ссылку.' },
        reaction: [],
        next_scene: null,
        finished: false,
        revision: 1,
      }),
    );
    const { result } = renderHook(() => useTrainingSession('scenario-1'));
    await waitFor(() => expect(result.current.training).not.toBeNull());
    await act(async () => {
      await result.current.chooseOption('safe');
    });
    expect(result.current.status).toBe('error');
    expect(result.current.errorMessage).toContain('следующую сцену');
  });

  it('redirects to login when choice is unauthorized', async () => {
    choose.mockReturnValueOnce(rejected({ status: 401 }));
    const { result } = renderHook(() => useTrainingSession('scenario-1'));
    await waitFor(() => expect(result.current.training).not.toBeNull());
    await act(async () => {
      await result.current.chooseOption('safe');
    });
    expect(navigate).toHaveBeenCalledWith('/login', { replace: true });
  });

  it('keeps the active scene after a generic choice error', async () => {
    choose.mockReturnValueOnce(rejected({ status: 500 }));
    const { result } = renderHook(() => useTrainingSession('scenario-1'));
    await waitFor(() => expect(result.current.training).not.toBeNull());
    await act(async () => {
      await result.current.chooseOption('safe');
    });
    expect(result.current.status).toBe('active');
    expect(result.current.errorMessage).toContain('Не удалось сохранить выбор');
    expect(result.current.messages.map(({ author, text }) => ({ author, text }))).toEqual([
      { author: 'counterpart', text: 'Начало' },
    ]);
  });

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
      void result.current.chooseOption('safe');
    });
    expect(result.current.status).toBe('submitting');
    await act(async () => {
      void result.current.chooseOption('safe');
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

  it('keeps the next scene hidden until feedback is acknowledged', async () => {
    choose.mockReturnValue(
      resolved({
        feedback: { verdict: 'risky', text: 'Проверьте адрес сайта.' },
        reaction: [
          { author: 'user', text: 'Проверю адрес сайта.' },
          { author: 'counterpart', text: 'Почему вы так долго?' },
        ],
        next_scene: secondScene,
        finished: false,
        revision: 1,
      }),
    );
    const { result } = renderHook(() => useTrainingSession('scenario-1'));
    await waitFor(() => expect(result.current.training).not.toBeNull());

    await act(async () => {
      await result.current.chooseOption('safe');
    });
    expect(result.current.status).toBe('feedback');
    expect(result.current.training?.scene.scene_id).toBe('scene-1');
    expect(result.current.messages.map(({ author, text }) => ({ author, text }))).toEqual([
      { author: 'counterpart', text: 'Начало' },
      { author: 'user', text: 'Проверю адрес сайта.' },
      { author: 'counterpart', text: 'Почему вы так долго?' },
    ]);
    expect(result.current.messages.at(-1)?.text).toBe('Почему вы так долго?');

    act(() => result.current.continueAfterFeedback());
    expect(result.current.training?.scene.scene_id).toBe('scene-2');
    expect(result.current.status).toBe('active');
  });

  it('recovers from a revision conflict by reloading the active attempt', async () => {
    choose.mockReturnValueOnce(rejected({ status: 409 }));
    const { result } = renderHook(() => useTrainingSession('scenario-1'));
    await waitFor(() => expect(result.current.training).not.toBeNull());

    await act(async () => {
      await result.current.chooseOption('safe');
    });
    await waitFor(() => expect(start).toHaveBeenCalledTimes(2));
    expect(result.current.status).toBe('active');
    expect(result.current.errorMessage).toBeNull();
  });

  it('ignores empty messages returned by the server', async () => {
    choose.mockReturnValueOnce(
      resolved({
        feedback: { verdict: 'safe', text: 'Готово.' },
        reaction: [
          { author: 'user', text: '  ' },
          { author: 'system', text: 'Профиль заблокирован.' },
        ],
        next_scene: null,
        finished: true,
        revision: 1,
      }),
    );
    const { result } = renderHook(() => useTrainingSession('scenario-1'));
    await waitFor(() => expect(result.current.training).not.toBeNull());

    await act(async () => {
      await result.current.chooseOption('safe');
    });

    expect(result.current.messages.map(({ author, text }) => ({ author, text }))).toEqual([
      { author: 'counterpart', text: 'Начало' },
      { author: 'system', text: 'Профиль заблокирован.' },
    ]);
  });
});
