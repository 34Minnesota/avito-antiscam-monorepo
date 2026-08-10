import { useGetMeQuery } from '@/entities/User';
import { LogoutButton } from '@/features/logout';
import { Avatar } from '@/shared/ui/Avatar';
import { Logo } from '@/shared/ui/Logo';
import cls from './AppHeader.module.scss';

export const AppHeader = () => {
  const { data: user } = useGetMeQuery();

  return (
    <header className={cls.Header}>
      <Logo />
      <div className={cls.right}>
        {user ? (
          <div className={cls.user}>
            <Avatar name={user.nickname} size="s" />
            <div>
              <b>{user.nickname}</b>
              <span>{user.email}</span>
            </div>
          </div>
        ) : null}
        <LogoutButton />
      </div>
    </header>
  );
};
