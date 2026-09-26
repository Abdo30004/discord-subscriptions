import { config } from '../config';

export interface Plan {
  id: string;
  name: string;
  description: string;
  interval: string;
  price_cents: number;
  is_dedicated: boolean;
  features: string[];
}

export interface BotTemplate {
  id: string;
  slug: string;
  name: string;
  description: string;
  category: string;
  supports_dedicated: boolean;
  plans?: Plan[];
}

export async function fetchBotCatalog(): Promise<BotTemplate[]> {
  try {
    const res = await fetch(`${config.catalogSvcUrl}/api/v1/catalog/bots`);
    if (!res.ok) {
      throw new Error(`Catalog service returned status ${res.status}`);
    }
    const data = await res.json();
    return data.data || [];
  } catch (err) {
    console.error('[API] Failed to fetch bot catalog:', err);
    return [];
  }
}
