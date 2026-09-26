import { SlashCommandBuilder, ChatInputCommandInteraction, EmbedBuilder, PermissionFlagsBits } from 'discord.js';
import { redeemVoucherCode } from '../api/billing';

export const data = new SlashCommandBuilder()
  .setName('redeem')
  .setDescription('Redeem a subscription gift card or voucher code for this server')
  .setDefaultMemberPermissions(PermissionFlagsBits.ManageGuild)
  .addStringOption(option =>
    option.setName('code')
      .setDescription('The gift or voucher code (e.g., GIFT-MUSIC-PRO-30D)')
      .setRequired(true)
  );

export async function execute(interaction: ChatInputCommandInteraction) {
  if (!interaction.guildId) {
    await interaction.reply({ content: 'This command can only be used inside a Discord server.', ephemeral: true });
    return;
  }

  const code = interaction.options.getString('code', true).trim();
  await interaction.deferReply({ ephemeral: true });

  const result = await redeemVoucherCode(code, interaction.user.id, interaction.guildId);

  if (!result.success) {
    const errorEmbed = new EmbedBuilder()
      .setTitle('❌ Voucher Redemption Failed')
      .setDescription(result.message)
      .setColor(0xed4245)
      .setTimestamp();

    await interaction.editReply({ embeds: [errorEmbed] });
    return;
  }

  const sub = result.subscription;
  const successEmbed = new EmbedBuilder()
    .setTitle('🎉 Subscription Activated!')
    .setDescription(`Successfully redeemed **${code}** for this server!`)
    .setColor(0x57f287)
    .addFields(
      { name: 'Bot Type', value: sub?.bot_type?.toUpperCase() || 'STANDARD', inline: true },
      { name: 'Hosting Mode', value: sub?.is_dedicated ? 'Dedicated K8s Instance' : 'Shared Cluster', inline: true },
      { name: 'Valid Until', value: sub?.valid_until ? new Date(sub.valid_until).toLocaleDateString() : 'Active', inline: true }
    )
    .setFooter({ text: 'Our deployment orchestrator is now preparing your bot container.' })
    .setTimestamp();

  await interaction.editReply({ embeds: [successEmbed] });
}
