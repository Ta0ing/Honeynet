import { useState } from 'react';
import { IconApps, IconSafe } from '@arco-design/web-react/icon';

type Props = {
  src?: string;
  alt: string;
  kind?: 'system' | 'company';
  className?: string;
};

export default function BrandLogo({ src, alt, kind = 'system', className = '' }: Props) {
  const [failedURL, setFailedURL] = useState('');
  const failed = !src || failedURL === src;
  return <span className={`brand-logo ${className}`.trim()} aria-label={alt}>
    {failed ? (kind === 'system' ? <IconApps /> : <IconSafe />) : <img src={src} alt={alt} onError={() => setFailedURL(src)} />}
  </span>;
}
