import React from 'react';
import { getCatalogBotsServer } from '@/lib/server-api';
import { StoreClient } from '@/components/store/StoreClient';
import { BotTemplate } from '@/lib/types';

export const revalidate = 60;

export default async function StorePage() {
  const initialBots: BotTemplate[] = await getCatalogBotsServer();

  return <StoreClient initialBots={initialBots} />;
}
