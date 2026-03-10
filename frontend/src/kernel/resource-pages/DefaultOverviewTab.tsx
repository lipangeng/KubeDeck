import type { ReactNode } from 'react';

interface DefaultOverviewTabProps {
  children?: ReactNode;
}

export function DefaultOverviewTab({ children }: DefaultOverviewTabProps) {
  return <>{children}</>;
}
