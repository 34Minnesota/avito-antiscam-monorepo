import { Component, ErrorInfo, ReactNode } from 'react';
import cls from './ErrorBoundary.module.scss';

interface Props {
  children: ReactNode;
}
interface State {
  hasError: boolean;
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('AntiScam UI error', error, info);
  }

  private reset = () => {
    window.location.reload();
  };

  render() {
    if (!this.state.hasError) return this.props.children;
    return (
      <main className={cls.page}>
        <section className={cls.card} role="alert">
          <div className={cls.icon}>!</div>
          <h1>Не удалось отобразить экран</h1>
          <p>
            Если последний шаг был сохранён сервером, его можно восстановить при повторном открытии
            тренировки. Попробуйте перезагрузить экран или вернуться к сценариям.
          </p>
          <div className={cls.actions}>
            <button className={cls.button} type="button" onClick={this.reset}>
              Перезагрузить экран
            </button>
            <button
              className={cls.button}
              type="button"
              onClick={() => window.location.assign('/')}
            >
              К сценариям
            </button>
          </div>
        </section>
      </main>
    );
  }
}
