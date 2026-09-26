import { Interaction } from 'discord.js';
import { commands } from '../commands';

export async function handleInteraction(interaction: Interaction) {
  if (!interaction.isChatInputCommand()) return;

  const command = commands.get(interaction.commandName);
  if (!command) {
    console.warn(`[Command] Unknown command invoked: ${interaction.commandName}`);
    return;
  }

  try {
    await command.execute(interaction);
  } catch (err: any) {
    console.error(`[Command Error] Error executing /${interaction.commandName}:`, err);
    const replyContent = { content: '⚠️ An unexpected error occurred while executing this command.', ephemeral: true };
    if (interaction.replied || interaction.deferred) {
      await interaction.followUp(replyContent);
    } else {
      await interaction.reply(replyContent);
    }
  }
}
