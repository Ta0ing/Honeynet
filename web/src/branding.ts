import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from './api';

export const DEFAULT_BRANDING = {
  system_name: 'Honeynet',
  system_version: '',
  copyright: '© Honeynet',
  customer_service_phone: '',
  customer_service_email: '',
  official_website_url: '',
  product_documentation_url: '',
  system_logo_url: '',
  company_logo_url: '',
  revision: 1,
} as const;

export type PlatformBranding = {
  system_name: string;
  system_version: string;
  copyright: string;
  customer_service_phone: string;
  customer_service_email: string;
  official_website_url: string;
  product_documentation_url: string;
  system_logo_url: string;
  company_logo_url: string;
  revision: number;
};

export type PlatformConfig = PlatformBranding & {
  updated_at?: string;
};

export const platformOEMQueryKey = ['platform-oem'] as const;

function cleanText(value: unknown, fallback = '') {
  return typeof value === 'string' ? value.trim() : fallback;
}

function safePublicURL(value: unknown) {
  const url = cleanText(value);
  if (!url) return '';
  if (url.startsWith('/') && !url.startsWith('//') && !url.startsWith('/\\')) return url;
  try {
    const parsed = new URL(url, window.location.origin);
    return parsed.protocol === 'http:' || parsed.protocol === 'https:' ? parsed.href : '';
  } catch {
    return '';
  }
}

export function normalizeBranding(value: any): PlatformBranding {
  return {
    system_name: cleanText(value?.system_name, DEFAULT_BRANDING.system_name) || DEFAULT_BRANDING.system_name,
    system_version: cleanText(value?.system_version),
    copyright: cleanText(value?.copyright, DEFAULT_BRANDING.copyright) || DEFAULT_BRANDING.copyright,
    customer_service_phone: cleanText(value?.customer_service_phone),
    customer_service_email: cleanText(value?.customer_service_email),
    official_website_url: safePublicURL(value?.official_website_url),
    product_documentation_url: safePublicURL(value?.product_documentation_url),
    system_logo_url: safePublicURL(value?.system_logo_url),
    company_logo_url: safePublicURL(value?.company_logo_url),
    revision: Number.isSafeInteger(value?.revision) && value.revision > 0 ? value.revision : DEFAULT_BRANDING.revision,
  };
}

async function fetchBranding() {
  try {
    return normalizeBranding(await api.get('/platform/branding'));
  } catch {
    return normalizeBranding(DEFAULT_BRANDING);
  }
}

export function usePlatformBranding() {
  return useQuery<PlatformBranding>({
    queryKey: platformOEMQueryKey,
    queryFn: fetchBranding,
    staleTime: 5 * 60_000,
    gcTime: 24 * 60 * 60_000,
    retry: false,
    placeholderData: normalizeBranding(DEFAULT_BRANDING),
  });
}

export function useApplyPlatformIdentity(branding: PlatformBranding | undefined) {
  useEffect(() => {
    const safe = normalizeBranding(branding);
    document.title = safe.system_name;
    if (!safe.system_logo_url) {
      document.querySelector<HTMLLinkElement>('link[data-platform-favicon="true"]')?.remove();
      return;
    }
    let cancelled = false;
    const candidate = new Image();
    candidate.onload = () => {
      if (cancelled) return;
      let favicon = document.querySelector<HTMLLinkElement>('link[data-platform-favicon="true"]');
      if (!favicon) {
        favicon = document.createElement('link');
        favicon.rel = 'icon';
        favicon.dataset.platformFavicon = 'true';
        document.head.appendChild(favicon);
      }
      favicon.href = safe.system_logo_url;
    };
    candidate.onerror = () => {
      if (!cancelled) document.querySelector<HTMLLinkElement>('link[data-platform-favicon="true"]')?.remove();
    };
    candidate.src = safe.system_logo_url;
    return () => { cancelled = true; };
  }, [branding]);
}

export function publicExternalURL(value: string) {
  return safePublicURL(value);
}
