import type { ReactNode } from 'react';

interface DefaultYamlTabProps {
  children?: ReactNode;
}

export function DefaultYamlTab({ children }: DefaultYamlTabProps) {
  return <>{children}</>;
}
