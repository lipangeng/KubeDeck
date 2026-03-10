import { useLocale } from './localeContext';
import { enMessages } from './messages/en';
import type { SupportedLocale } from './types';

type Catalog = typeof enMessages;

const catalogs: Record<SupportedLocale, Catalog> = {
  en: enMessages,
};

export type MessageKey = keyof Catalog;

export function copy(key: MessageKey, locale: SupportedLocale = 'en'): string {
  return catalogs[locale][key];
}

export function useCopy(key: MessageKey): string {
  const { locale } = useLocale();
  return copy(key, locale.current);
}
