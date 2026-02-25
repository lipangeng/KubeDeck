import type { PageContribution } from '../sdk/types';

export function resolvePage(
  pages: PageContribution[],
  pageId: string,
): PageContribution | undefined {
  const replacement = pages.find((page) => page.replacementFor === pageId);
  if (replacement) {
    return replacement;
  }

  return pages.find((page) => page.pageId === pageId);
}
