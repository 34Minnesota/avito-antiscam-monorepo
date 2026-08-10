import cls from './Logo.module.scss';
export const Logo = () => (
  <div className={cls.Logo}>
    <div className={cls.mark}>A</div>
    <div>
      <div className={cls.title}>AntiScam</div>
      <div className={cls.caption}>ТРЕНАЖЁР БЕЗОПАСНЫХ СДЕЛОК</div>
    </div>
  </div>
);
