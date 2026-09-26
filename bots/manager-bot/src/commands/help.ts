import { SlashCommandBuilder, ChatInputCommandInteraction, EmbedBuilder } from 'discord.js';
import { config } from '../config';

export const data = new SlashCommandBuilder()
  .setName('help')
  .setDescription('List all available Manager Bot commands and usage information');

export async function execute(interaction: ChatInputCommandInteraction) {
  const embed = new EmbedBuilder()
    .setTitle('📖 Manager Bot Help & Command Center')
    .setDescription('Manage your server bot subscriptions, check live uptime metrics, and control deployments directly from Discord.')
    .setColor(0x5865f2)
    .addFields(
      { name: '🛒 `/store`', value: 'Browse available bots, subscription tiers, and features.' },
      { name: '🎟️ `/redeem <code>`', value: 'Claim a gift card or voucher code to activate a subscription instantly.' },
      { name: '📊 `/status`', value: 'View live pod health status, WebSocket latency, memory usage, and expiration.' },
      { name: '🔄 `/restart`', value: 'Trigger a zero-downtime rolling restart of your dedicated bot pod.' },
      { name: '🎨 `/bot name <name>` / `/bot avatar <url>`', value: 'Customize your bot username and profile picture.' },
      { name: '🌐 Web Dashboard', value: `Configure custom bot tokens and preferences at: [Dashboard](${config.dashboardUrl})` }
    )
    .setFooter({ text: 'Discord Subscriptions Platform' })
    .setTimestamp();

  await interaction.reply({ embeds: [embed] });
}
