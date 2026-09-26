import { Client, ActivityType } from 'discord.js';

export function handleReady(client: Client<true>) {
  console.log(`[Bot Ready] Logged in as ${client.user.tag} (ID: ${client.user.id})`);
  client.user.setActivity({
    name: 'Bot Subscriptions | /help',
    type: ActivityType.Custom,
  });
}
