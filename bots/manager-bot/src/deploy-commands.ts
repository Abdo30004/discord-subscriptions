import { REST, Routes } from 'discord.js';
import { config } from './config';
import { commandData } from './commands';

export async function deployCommands() {
  if (!config.discordToken || !config.clientId) {
    console.warn('[Deploy Commands] Skipping command deployment: MANAGER_BOT_TOKEN or MANAGER_BOT_CLIENT_ID not set.');
    return;
  }

  const rest = new REST({ version: '10' }).setToken(config.discordToken);

  try {
    console.log(`[Deploy Commands] Registering ${commandData.length} application (/) commands...`);
    await rest.put(
      Routes.applicationCommands(config.clientId),
      { body: commandData }
    );
    console.log('[Deploy Commands] Successfully reloaded application (/) commands globally.');
  } catch (error) {
    console.error('[Deploy Commands] Failed to register slash commands:', error);
  }
}

if (require.main === module) {
  deployCommands();
}
