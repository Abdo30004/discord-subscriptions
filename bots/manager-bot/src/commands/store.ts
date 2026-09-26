import { SlashCommandBuilder, ChatInputCommandInteraction, EmbedBuilder, ActionRowBuilder, ButtonBuilder, ButtonStyle } from 'discord.js';
import { fetchBotCatalog } from '../api/catalog';
import { config } from '../config';

export const data = new SlashCommandBuilder()
  .setName('store')
  .setDescription('Browse available Discord bots and subscription tiers');

export async function execute(interaction: ChatInputCommandInteraction) {
  await interaction.deferReply();

  const bots = await fetchBotCatalog();
  if (bots.length === 0) {
    await interaction.editReply({ content: '⚠️ The store catalog is currently unavailable. Please try again shortly.' });
    return;
  }

  const embed = new EmbedBuilder()
    .setTitle('🤖 Discord Bot Subscription Store')
    .setDescription('Select and subscribe to custom-branded bots deployed on our high-availability cloud infrastructure.')
    .setColor(0x5865f2)
    .setTimestamp();

  for (const bot of bots) {
    let plansText = '';
    if (bot.plans && bot.plans.length > 0) {
      plansText = bot.plans
        .map(p => `• **${p.name}**: ${p.price_cents === 0 ? 'Free' : `$${(p.price_cents / 100).toFixed(2)}/mo`} ${p.is_dedicated ? '*(Dedicated Pod)*' : '*(Shared)*'}`)
        .join('\n');
    } else {
      plansText = 'No active pricing tiers.';
    }

    embed.addFields({
      name: `${bot.name} (${bot.category.toUpperCase()})`,
      value: `${bot.description}\n\n**Plans:**\n${plansText}`,
      inline: false,
    });
  }

  const row = new ActionRowBuilder<ButtonBuilder>().addComponents(
    new ButtonBuilder()
      .setLabel('Visit Web Dashboard')
      .setURL(config.dashboardUrl)
      .setStyle(ButtonStyle.Link)
  );

  await interaction.editReply({ embeds: [embed], components: [row] });
}
