import { ReactNode } from 'react';
import cls from './Badge.module.scss';
export const Badge = ({
  children,
  tone = 'neutral',
}: {
  children: ReactNode;
  tone?: 'neutral' | 'good' | 'warn' | 'danger' | 'accent';
}) => <span className={`${cls.Badge} ${cls[tone]}`}>{children}</span>;
