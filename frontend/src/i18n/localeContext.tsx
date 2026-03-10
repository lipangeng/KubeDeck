import {
  createContext,
  type PropsWithChildren,
  useContext,
  useMemo,
  useState,
} from 'react';
import type { LocaleState, SupportedLocale } from './types';

interface LocaleContextValue {
  locale: LocaleState;
  setLocale: (next: SupportedLocale) => void;
}

const LocaleContext = createContext<LocaleContextValue | null>(null);

interface LocaleProviderProps extends PropsWithChildren {
  initialLocale?: SupportedLocale;
}

export function LocaleProvider({ children, initialLocale = 'en' }: LocaleProviderProps) {
  const [current, setCurrent] = useState<SupportedLocale>(initialLocale);
  const value = useMemo<LocaleContextValue>(
    () => ({
      locale: { current },
      setLocale: setCurrent,
    }),
    [current],
  );

  return <LocaleContext.Provider value={value}>{children}</LocaleContext.Provider>;
}

export function useLocale(): LocaleContextValue {
  const context = useContext(LocaleContext);
  if (!context) {
    throw new Error('useLocale must be used within LocaleProvider');
  }
  return context;
}
