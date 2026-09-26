import {
  SlashCommandBuilder,
  ChatInputCommandInteraction,
  EmbedBuilder,
  ActionRowBuilder,
  StringSelectMenuBuilder,
  StringSelectMenuOptionBuilder,
  ComponentType,
} from 'discord.js';
import { getGuildSubscriptions, Subscription } from '../api/billing';
import { getGuildDeployments, Deployment } from '../api/deploy';
import { getGuildLiveStatus, GuildStatusResponse } from '../api/monitor';

export const data = new SlashCommandBuilder()
  .setName('status')
  .setDescription('Check the live operational status and subscription details for this server');

function buildDetailEmbed(
  guildName: string,
  dep?: Deployment,
  sub?: Subscription,
  monitor?: GuildStatusResponse | null
): EmbedBuilder {
  const matchingTarget = monitor?.targets?.find((t) => t.bot_id === dep?.id) || monitor?.target;
  const liveStatus = matchingTarget?.current_status || dep?.status || 'unknown';
  const statusColor = liveStatus === 'online' ? 0x57f287 : liveStatus === 'degraded' ? 0xfee75c : 0xed4245;
  const latestLog = monitor?.recent_logs && monitor.recent_logs.length > 0 ? monitor.recent_logs[0] : null;

  const instanceTitle = dep?.instance_label ? `${dep.instance_label} (${dep.bot_type.toUpperCase()})` : (sub?.bot_type?.toUpperCase() || 'Bot');

  return new EmbedBuilder()
    .setTitle(`🛡️ Bot Status: ${instanceTitle} - ${guildName}`)
    .setColor(statusColor)
    .addFields(
      { name: 'Instance Label', value: dep?.instance_label || sub?.instance_label || 'Default', inline: true },
      { name: 'Bot Type', value: sub?.bot_type?.toUpperCase() || dep?.bot_type?.toUpperCase() || 'N/A', inline: true },
      { name: 'Status', value: `● **${liveStatus.toUpperCase()}**`, inline: true },
      { name: 'Hosting Type', value: sub?.is_dedicated ? 'Dedicated K8s Pod' : 'Shared Multi-tenant', inline: true },
      { name: 'Subscription Expiry', value: sub?.valid_until ? new Date(sub.valid_until).toLocaleDateString() : 'N/A', inline: true },
      { name: 'Deployment ID', value: dep?.id || 'N/A', inline: true },
      { name: 'Pod Image Tag', value: dep?.image_tag || 'latest', inline: true },
      { name: 'Gateway Latency', value: latestLog?.latency_ms ? `${latestLog.latency_ms} ms` : 'N/A', inline: true },
      { name: 'Discord WebSocket Ping', value: latestLog?.discord_ping_ms ? `${latestLog.discord_ping_ms} ms` : 'N/A', inline: true }
    )
    .setFooter({ text: 'Monitored continuously by monitor-svc' })
    .setTimestamp();
}

export async function execute(interaction: ChatInputCommandInteraction) {
  if (!interaction.guildId) {
    await interaction.reply({ content: 'This command can only be used inside a server.', ephemeral: true });
    return;
  }

  await interaction.deferReply();
  const guildId = interaction.guildId;
  const guildName = interaction.guild?.name || 'Current Guild';

  const [subs, deps, monitor] = await Promise.all([
    getGuildSubscriptions(guildId),
    getGuildDeployments(guildId),
    getGuildLiveStatus(guildId),
  ]);

  if (subs.length === 0 && deps.length === 0) {
    const noSubEmbed = new EmbedBuilder()
      .setTitle('ℹ️ No Active Subscription')
      .setDescription('This server does not have an active bot subscription yet. Use `/store` to browse available bots!')
      .setColor(0xfee75c);

    await interaction.editReply({ embeds: [noSubEmbed] });
    return;
  }

  // If exactly 1 deployment (or 0 deployments but 1 sub), render directly
  if (deps.length <= 1) {
    const dep = deps[0];
    const sub = subs.find((s) => s.id === dep?.subscription_id) || subs[0];
    const embed = buildDetailEmbed(guildName, dep, sub, monitor);
    await interaction.editReply({ embeds: [embed] });
    return;
  }

  // Multi-bot fleet! Render fleet overview and interactive select menu
  const fleetEmbed = new EmbedBuilder()
    .setTitle(`🛡️ Server Bot Fleet: ${guildName}`)
    .setDescription(
      `This server is running **${deps.length} active bot instances**. Select an instance from the menu below to view detailed operational telemetry.`
    )
    .setColor(0x5865f2);

  deps.forEach((d, idx) => {
    const target = monitor?.targets?.find((t) => t.bot_id === d.id);
    const st = target?.current_status || d.status || 'unknown';
    const label = d.instance_label || `Instance #${idx + 1}`;
    fleetEmbed.addFields({
      name: `${idx + 1}. ${label} [${d.bot_type.toUpperCase()}]`,
      value: `Status: ● **${st.toUpperCase()}** | Pod ID: \`${d.id}\``,
      inline: false,
    });
  });

  const selectMenu = new StringSelectMenuBuilder()
    .setCustomId('select_status_bot')
    .setPlaceholder('🔍 Select a bot instance to inspect...')
    .addOptions(
      deps.map((d, idx) =>
        new StringSelectMenuOptionBuilder()
          .setLabel(d.instance_label || `${d.bot_type} #${idx + 1}`)
          .setDescription(`Type: ${d.bot_type.toUpperCase()} | Status: ${d.status}`)
          .setValue(d.id)
      )
    );

  const row = new ActionRowBuilder<StringSelectMenuBuilder>().addComponents(selectMenu);
  const reply = await interaction.editReply({ embeds: [fleetEmbed], components: [row] });

  const collector = reply.createMessageComponentCollector({
    componentType: ComponentType.StringSelect,
    time: 120_000,
  });

  collector.on('collect', async (i) => {
    if (i.user.id !== interaction.user.id) {
      await i.reply({ content: 'Only the user who ran /status can change the selected bot.', ephemeral: true });
      return;
    }

    const selectedDepId = i.values[0];
    const selectedDep = deps.find((d) => d.id === selectedDepId);
    const matchingSub = subs.find((s) => s.id === selectedDep?.subscription_id);
    const detailEmbed = buildDetailEmbed(guildName, selectedDep, matchingSub, monitor);

    await i.update({ embeds: [detailEmbed], components: [row] });
  });

  collector.on('end', async () => {
    // Disable menu on timeout
    try {
      selectMenu.setDisabled(true);
      await interaction.editReply({ components: [row] });
    } catch {
      // Message may have been deleted
    }
  });
}
