import { FormEvent, useId, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLoginMutation, useRegisterMutation } from '@/entities/User';
import { setSessionId } from '@/shared/auth/session';
import { getApiMessage } from '@/shared/api/errors';
import { Button } from '@/shared/ui/Button';
import { Input } from '@/shared/ui/Input';
import { Logo } from '@/shared/ui/Logo';
import cls from './AuthPage.module.scss';

export const AuthPage = () => {
  const navigate = useNavigate();
  const tabId = useId();
  const [register, setRegister] = useState(false);
  const [nickname, setNickname] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [login, { isLoading: loginLoading, error: loginError }] = useLoginMutation();
  const [signup, { isLoading: signupLoading, error: signupError }] = useRegisterMutation();

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    try {
      const result = register
        ? await signup({ nickname, email, password }).unwrap()
        : await login({ email, password }).unwrap();
      setSessionId(result.sessionId);
      navigate('/', { replace: true });
    } catch {
      // Mutation state renders the server error.
    }
  };

  const loading = loginLoading || signupLoading;
  const error = register ? signupError : loginError;
  const panelId = `${tabId}-panel`;

  return (
    <div className={cls.page}>
      <div className={cls.glow} />
      <div className={cls.wrap}>
        <div className={cls.brand}>
          <Logo />
          <div className={cls.pitch}>
            <span>Безопасность</span> — это навык, который можно натренировать.
          </div>
          <div className={cls.points}>
            <div>
              <b>01</b>
              <span>Реалистичные ситуации сделок</span>
            </div>
            <div>
              <b>02</b>
              <span>Последствия каждого решения</span>
            </div>
            <div>
              <b>03</b>
              <span>Разбор и рост результата</span>
            </div>
          </div>
        </div>

        <div className={cls.formCard}>
          <div className={cls.tabs} role="tablist" aria-label="Авторизация">
            <button
              type="button"
              role="tab"
              aria-selected={!register}
              aria-controls={panelId}
              id={`${tabId}-login`}
              className={!register ? cls.active : ''}
              onClick={() => setRegister(false)}
            >
              Войти
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={register}
              aria-controls={panelId}
              id={`${tabId}-register`}
              className={register ? cls.active : ''}
              onClick={() => setRegister(true)}
            >
              Регистрация
            </button>
          </div>
          <div
            id={panelId}
            role="tabpanel"
            aria-labelledby={register ? `${tabId}-register` : `${tabId}-login`}
          >
            <h1>{register ? 'Создайте аккаунт' : 'С возвращением'}</h1>
            <p>
              {register
                ? 'Сохраните прогресс и тренируйтесь в обеих ролях.'
                : 'Продолжите с того места, где остановились.'}
            </p>
            <form onSubmit={submit}>
              {register && (
                <label className={cls.field}>
                  <span>Имя на платформе</span>
                  <Input
                    placeholder="Например, mops"
                    value={nickname}
                    onChange={(event) => setNickname(event.target.value)}
                    minLength={3}
                    maxLength={30}
                    required
                  />
                </label>
              )}
              <label className={cls.field}>
                <span>Email</span>
                <Input
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  required
                />
              </label>
              <label className={cls.field}>
                <span>Пароль</span>
                <Input
                  type="password"
                  placeholder="Минимум 8 символов"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  minLength={8}
                  required
                />
              </label>
              {error && (
                <div className={cls.error} role="alert">
                  {getApiMessage(error, 'Не удалось выполнить запрос')}
                </div>
              )}
              <Button type="submit" size="l" fullWidth disabled={loading}>
                {loading ? 'Проверяем…' : register ? 'Создать аккаунт →' : 'Войти →'}
              </Button>
            </form>
            <div className={cls.hint}>
              Прогресс хранится на сервере. Session ID используется только для доступа к вашим
              тренировкам.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
