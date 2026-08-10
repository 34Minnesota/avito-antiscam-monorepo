import { InputHTMLAttributes } from 'react';
import { classNames } from '@/shared/lib/classNames';
import cls from './Input.module.scss';
export const Input = ({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) => (
  <input className={classNames(cls.Input, className)} {...props} />
);
