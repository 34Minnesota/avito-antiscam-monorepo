import { useNavigate } from 'react-router-dom';
import { Button } from '@/shared/ui/Button';
import cls from './NotFoundPage.module.scss';
export const NotFoundPage = () => {
  const navigate = useNavigate();
  return (
    <main className={cls.page}>
      <div>
        <span>404</span>
        <h1>Страница не найдена</h1>
        <p>Похоже, вы свернули не туда.</p>
        <Button onClick={() => navigate('/')}>Вернуться в AntiScam</Button>
      </div>
    </main>
  );
};
