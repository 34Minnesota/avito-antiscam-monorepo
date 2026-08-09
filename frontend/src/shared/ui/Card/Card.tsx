import { HTMLAttributes, ReactNode } from 'react';
import { classNames } from '@/shared/lib/classNames';
import cls from './Card.module.scss';
export const Card = ({
  className,
  children,
  ...props
}: HTMLAttributes<HTMLDivElement> & { children: ReactNode }) => (
  <div className={classNames(cls.Card, className)} {...props}>
    {children}
  </div>
);
