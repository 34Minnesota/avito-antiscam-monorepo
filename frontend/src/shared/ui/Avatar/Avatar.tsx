import cls from './Avatar.module.scss';
export const Avatar = ({ name, size = 'm' }: { name: string; size?: 's' | 'm' | 'l' }) => (
  <div className={`${cls.Avatar} ${cls[size]}`}>{name.slice(0, 1).toUpperCase()}</div>
);
