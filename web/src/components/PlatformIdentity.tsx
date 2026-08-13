import type { PropsWithChildren } from 'react';
import { useApplyPlatformIdentity, usePlatformBranding } from '../branding';

export default function PlatformIdentity({ children }: PropsWithChildren) {
  const branding = usePlatformBranding();
  useApplyPlatformIdentity(branding.data);
  return children;
}
