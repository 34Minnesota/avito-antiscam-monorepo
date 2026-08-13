import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import {
  useStartAttemptMutation,
  useSubmitChoiceMutation,
} from '@/features/training/api';
import { clearSessionId, getSessionId } from '@/shared/auth/session';
import {
  ChoiceResult,
  Message,
  StartResult,
  Verdict,
} from '@/shared/api/contracts';
import { isConflict, isUnauthorized } from '@/shared/api/errors';
import {
  getAttemptSnapshot,
  saveAttemptScenario,
  saveAttemptSnapshot,
} from './storage';
import { withStableMessageIds } from '@/shared/lib/chat/messageIds';

export type TrainingStatus =
  | 'starting'
  | 'active'
  | 'feedback'
  | 'submitting'
  | 'typing'
  | 'finished'
  | 'error';

export type TrainingFeedback = {
  verdict: Verdict;
  text: string;
};

const TYPING_DELAY_MIN_MS = 800;
const TYPING_DELAY_RANGE_MS = 1000;

export const getCounterpartTypingDelay = (
  random: () => number = Math.random,
) => {
  const normalizedRandom = Math.min(1, Math.max(0, random()));

  return Math.round(
    TYPING_DELAY_MIN_MS + normalizedRandom * TYPING_DELAY_RANGE_MS,
  );
};

export const useTrainingSession = (scenarioId?: string) => {
  const navigate = useNavigate();

  const [start] = useStartAttemptMutation();
  const [choose] = useSubmitChoiceMutation();

  const [training, setTraining] = useState<StartResult | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [sceneNumber, setSceneNumber] = useState(1);
  const [feedback, setFeedback] = useState<TrainingFeedback | null>(null);
  const [pendingScene, setPendingScene] = useState<
    StartResult['scene'] | null
  >(null);
  const [status, setStatus] = useState<TrainingStatus>('starting');
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const startedFor = useRef<string | null>(null);
  const typingTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const resolveTyping = useRef<((active: boolean) => void) | null>(null);
  const typingRun = useRef(0);

  const cancelTyping = useCallback(() => {
    typingRun.current += 1;

    if (typingTimer.current) {
      clearTimeout(typingTimer.current);
      typingTimer.current = null;
    }

    resolveTyping.current?.(false);
    resolveTyping.current = null;
  }, []);

  const waitForCounterpart = useCallback(() => {
    const run = ++typingRun.current;

    return new Promise<boolean>((resolve) => {
      resolveTyping.current = resolve;

      typingTimer.current = setTimeout(() => {
        typingTimer.current = null;
        resolveTyping.current = null;

        resolve(run === typingRun.current);
      }, getCounterpartTypingDelay());
    });
  }, []);

  useEffect(() => cancelTyping, [cancelTyping, scenarioId]);

  const restore = useCallback(
    (result: StartResult) => {
      const snapshot = getAttemptSnapshot(result.attempt_id);

      const resumedFromPreviousScene = Boolean(
        snapshot?.sceneId &&
          snapshot.sceneId !== result.scene.scene_id,
      );

      const nextMessages = resumedFromPreviousScene
        ? [...(snapshot?.messages ?? []), ...(result.scene.intro ?? [])]
        : snapshot?.messages?.length
          ? snapshot.messages
          : (result.scene?.intro ?? []);

      const nextSceneNumber = resumedFromPreviousScene
        ? (snapshot?.sceneNumber ?? 1) + 1
        : (snapshot?.sceneNumber ?? 1);

      const normalizedMessages = withStableMessageIds(
        nextMessages,
        result.attempt_id,
      );

      setTraining(result);
      setMessages(normalizedMessages);
      setSceneNumber(nextSceneNumber);
      setFeedback(null);
      setPendingScene(null);
      setErrorMessage(null);
      setStatus('active');

      saveAttemptScenario(
        result.attempt_id,
        scenarioId ?? '',
      );

      saveAttemptSnapshot(result.attempt_id, {
        sceneNumber: nextSceneNumber,
        sceneId: result.scene.scene_id,
        messages: normalizedMessages,
        revision: result.revision ?? snapshot?.revision ?? 0,
      });
    },
    [scenarioId],
  );

  const startSession = useCallback(async () => {
    if (!scenarioId || !getSessionId()) {
      return;
    }

    cancelTyping();
    setStatus('starting');
    setErrorMessage(null);

    try {
      const result = await start(scenarioId).unwrap();

      restore(result);
    } catch (error) {
      if (isUnauthorized(error)) {
        clearSessionId();
        navigate('/login', { replace: true });
        return;
      }

      setStatus('error');
      setErrorMessage(
        'Не удалось открыть сценарий. Вернитесь к каталогу и попробуйте ещё раз.',
      );
    }
  }, [
    cancelTyping,
    navigate,
    restore,
    scenarioId,
    start,
  ]);

  useEffect(() => {
    if (!scenarioId || !getSessionId()) {
      return;
    }

    if (startedFor.current === scenarioId) {
      return;
    }

    startedFor.current = scenarioId;

    void startSession();
  }, [scenarioId, startSession]);

  const chooseOption = useCallback(
    async (optionId: string) => {
      if (!training || status !== 'active') {
        return;
      }

      const currentAttemptId = training.attempt_id;

      setStatus('submitting');
      setErrorMessage(null);

      try {
        const result: ChoiceResult = await choose({
          attemptId: currentAttemptId,
          sceneId: training.scene.scene_id,
          optionId,
          expectedRevision:
            getAttemptSnapshot(currentAttemptId)?.revision ?? 0,
        }).unwrap();

        const reaction = (result.reaction ?? []).filter(
          (message) => message.text.trim().length > 0,
        );

        const withReaction = withStableMessageIds(
          [...messages, ...reaction],
          currentAttemptId,
        );

        const firstCounterpartIndex = reaction.findIndex(
          (message) => message.author === 'counterpart',
        );

        saveAttemptSnapshot(currentAttemptId, {
          sceneNumber,
          sceneId: training.scene.scene_id,
          messages: withReaction,
          revision: result.revision,
        });

        if (firstCounterpartIndex >= 0) {
          setMessages(
            withReaction.slice(
              0,
              messages.length + firstCounterpartIndex,
            ),
          );

          setFeedback(null);
          setStatus('typing');

          const active = await waitForCounterpart();

          if (!active) {
            return;
          }
        }

        setMessages(withReaction);
        setFeedback(result.feedback);

        if (result.finished) {
          setStatus('finished');
          return;
        }

        if (!result.next_scene) {
          setStatus('error');
          setErrorMessage(
            'Сервер не вернул следующую сцену. Попробуйте восстановить тренировку.',
          );
          return;
        }

        setPendingScene(result.next_scene);
        setStatus('feedback');
      } catch (error) {
        if (isUnauthorized(error)) {
          clearSessionId();
          navigate('/login', { replace: true });
          return;
        }

        if (isConflict(error)) {
          await startSession();
          return;
        }

        setStatus('active');
        setErrorMessage(
          'Не удалось сохранить выбор. Проверьте соединение и попробуйте ещё раз.',
        );
      }
    },
    [
      choose,
      messages,
      navigate,
      sceneNumber,
      startSession,
      status,
      training,
      waitForCounterpart,
    ],
  );

  const continueAfterFeedback = useCallback(() => {
    if (status === 'finished' && training) {
      navigate(`/result/${training.attempt_id}`, {
        replace: true,
      });

      return;
    }

    if (!pendingScene || !training) {
      return;
    }

    const nextNumber = sceneNumber + 1;

    const nextMessages = withStableMessageIds(
      [...messages, ...(pendingScene.intro ?? [])],
      training.attempt_id,
    );

    setTraining((current) =>
      current
        ? {
            ...current,
            scene: pendingScene,
          }
        : current,
    );

    setMessages(nextMessages);
    setSceneNumber(nextNumber);
    setPendingScene(null);
    setFeedback(null);
    setStatus('active');

    saveAttemptSnapshot(training.attempt_id, {
      sceneNumber: nextNumber,
      sceneId: pendingScene.scene_id,
      messages: nextMessages,
      revision:
        getAttemptSnapshot(training.attempt_id)?.revision ?? 0,
    });
  }, [
    messages,
    navigate,
    pendingScene,
    sceneNumber,
    status,
    training,
  ]);

  return {
    training,
    messages,
    sceneNumber,
    feedback,
    status,
    isCounterpartTyping: status === 'typing',
    errorMessage,
    chooseOption,
    continueAfterFeedback,
    retryStart: startSession,
  };
};