import { BotTemplate } from './types';

/**
 * Server-side data fetcher for Bot Templates from catalog-svc.
 * Designed for Next.js React Server Components with 60-second ISR caching.
 */
export async function getCatalogBotsServer(): Promise<BotTemplate[]> {
  const urlsToTry: string[] = [];

  if (process.env.INTERNAL_CATALOG_SVC_URL) {
    urlsToTry.push(`${process.env.INTERNAL_CATALOG_SVC_URL}/api/v1/bots`);
  }
  if (process.env.CATALOG_SVC_URL) {
    urlsToTry.push(`${process.env.CATALOG_SVC_URL}/api/v1/bots`);
  }

  // Internal Docker & Kubernetes DNS names
  urlsToTry.push('http://catalog-svc:8081/api/v1/bots');
  urlsToTry.push('http://traefik/api/v1/bots');
  urlsToTry.push('http://localhost:8081/api/v1/bots');
  urlsToTry.push('http://localhost/api/v1/bots');

  for (const url of urlsToTry) {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 2000);

      const res = await fetch(url, {
        next: { revalidate: 60 },
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (res.ok) {
        const body = await res.json();
        return body.data || body || [];
      }
    } catch {
      // Gracefully try fallback address
    }
  }

  return [];
}
