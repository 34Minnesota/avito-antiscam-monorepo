import { Role } from '@/shared/api/contracts';
import { classNames } from '@/shared/lib/classNames';
import cls from './RoleSwitcher.module.scss';

export const RoleSwitcher = ({
  value,
  onChange,
}: {
  value: Role;
  onChange: (role: Role) => void;
}) => (
  <div className={cls.Switcher} role="radiogroup" aria-label="Роль тренировки">
    {(['seller', 'buyer'] as Role[]).map((role) => (
      <button
        type="button"
        role="radio"
        aria-checked={value === role}
        key={role}
        className={classNames(cls.item, value === role && cls.active)}
        onClick={() => onChange(role)}
      >
        {role === 'seller' ? 'Я продавец' : 'Я покупатель'}
      </button>
    ))}
  </div>
);
