import { useNavigate, useParams } from 'react-router-dom';
import { Avatar } from '@/shared/ui/Avatar';
import { Badge } from '@/shared/ui/Badge';
import { Card } from '@/shared/ui/Card';
import { Loader } from '@/shared/ui/Loader';
import { AppHeader } from '@/widgets/AppHeader/AppHeader';
import { TrainingChat } from '@/widgets/TrainingChat/TrainingChat';
import { TrainingDecision } from '@/widgets/TrainingDecision/TrainingDecision';
import { DecisionFeedback } from '@/widgets/DecisionFeedback/DecisionFeedback';
import { useTrainingSession } from '@/features/training/model/useTrainingSession';
import { formatPrice } from '@/shared/lib/format/formatPrice';
import cls from './TrainingPage.module.scss';

export const TrainingPage = () => {
  const { scenarioId } = useParams();
  const navigate = useNavigate();
  const session = useTrainingSession(scenarioId);

  if (session.status === 'starting' || !session.training) {
    return (
      <>
        <AppHeader />
        <main className={cls.loading} aria-live="polite">
          {session.status === 'error' ? (
            <div className={cls.error} role="alert">
              <strong>{session.errorMessage}</strong>
              <button type="button" onClick={() => navigate('/')} className={cls.errorButton}>
                Вернуться к сценариям
              </button>
            </div>
          ) : (
            <Loader />
          )}
        </main>
      </>
    );
  }

  const { training, messages, feedback, sceneNumber, status, errorMessage } = session;
  const isSeller = training.role === 'seller';
  const listingImage = training.listing.image;

  return (
    <>
      <AppHeader />
      <main className={cls.page}>
        <div className={cls.topbar}>
          <button type="button" onClick={() => navigate('/')} className={cls.back}>
            ← Сценарии
          </button>
          <div
            className={cls.progress}
            aria-label={`Шаг ${sceneNumber} из ${training.scenes_total}`}
          >
            Шаг {sceneNumber} <span>/ {training.scenes_total}</span>
          </div>
        </div>

        {errorMessage && (
          <div className={cls.inlineError} role="alert">
            <span>{errorMessage}</span>
            <button type="button" onClick={() => void session.retryStart()}>
              Восстановить
            </button>
          </div>
        )}

        <div className={cls.layout}>
          <aside>
            <Card className={cls.listing}>
              <div className={cls.photo}>
                {listingImage ? (
                  <img src={listingImage} alt="" />
                ) : (
                  <span aria-hidden="true">{isSeller ? '📦' : '🛍️'}</span>
                )}
              </div>
              <Badge tone="accent">{isSeller ? 'Вы продавец' : 'Вы покупатель'}</Badge>
              <h2>{training.listing.title}</h2>
              <strong>{formatPrice(training.listing.price)} ₽</strong>
              <p>{training.listing.location}</p>
            </Card>

            <Card className={cls.person}>
              <div className={cls.personHead}>
                <Avatar name={training.counterpart.name} size="m" />
                <div>
                  <b>{training.counterpart.name}</b>
                  <span>
                    ★ {training.counterpart.rating.toFixed(1)} · {training.counterpart.reviews}{' '}
                    отзывов
                  </span>
                </div>
              </div>
              <div className={cls.personMeta}>{training.counterpart.registered}</div>
            </Card>

            <div className={cls.tip}>
              <b>Правило тренировки</b>
              <p>Не спешите. Проверяйте факты, даже если собеседник создаёт ощущение срочности.</p>
            </div>
          </aside>

          <section className={cls.chatCard} aria-label="Тренировочный чат">
            <div className={cls.chatHeader}>
              <div>
                <b>{training.counterpart.name}</b>
                <span>Онлайн · чат сделки</span>
              </div>
              <span className={cls.secure}>● защищённая тренировка</span>
            </div>
            <TrainingChat messages={messages} counterpartName={training.counterpart.name} />
            {feedback ? (
              <DecisionFeedback
                verdict={feedback.verdict}
                text={feedback.text}
                isFinal={status === 'finished'}
                onContinue={session.continueAfterFeedback}
              />
            ) : (
              <TrainingDecision
                prompt={training.scene.prompt}
                options={training.scene.options ?? []}
                onChoose={session.chooseOption}
                disabled={status === 'submitting' || Boolean(errorMessage)}
              />
            )}
          </section>
        </div>
      </main>
    </>
  );
};
