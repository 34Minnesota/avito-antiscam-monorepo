import { useState } from 'react';
import cls from './ImageWithFallback.module.scss';

interface Props {
  src?: string | null;
  alt: string;
  fallback: string;
  className?: string;
}

export const ImageWithFallback = ({ src, alt, fallback, className }: Props) => {
  const [failed, setFailed] = useState(!src);
  if (failed)
    return (
      <div className={`${cls.fallback} ${className ?? ''}`} aria-hidden="true">
        {fallback}
      </div>
    );
  return <img className={className} src={src ?? ''} alt={alt} onError={() => setFailed(true)} />;
};
