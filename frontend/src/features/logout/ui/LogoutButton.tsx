import { useLogoutMutation } from '@/entities/User';
import { clearSessionId } from '@/shared/auth/session';
import { Button } from '@/shared/ui/Button';

export const LogoutButton = () => {
  const [logout, { isLoading }] = useLogoutMutation();

  const onClick = async () => {
    try {
      await logout().unwrap();
    } finally {
      clearSessionId();
      window.location.assign('/login');
    }
  };

  return (
    <Button variant="ghost" onClick={onClick} disabled={isLoading}>
      Выйти
    </Button>
  );
};
