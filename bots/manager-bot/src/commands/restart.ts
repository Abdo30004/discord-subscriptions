import {
  SlashCommandBuilder,
  ChatInputCommandInteraction,
  EmbedBuilder,
  PermissionFlagsBits,
  ActionRowBuilder,
  StringSelectMenuBuilder,
  StringSelectMenuOptionBuilder,
  ComponentType,
} from 'discord.js';
import { getGuildDeployments, restartDeployment, Deployment } from '../api/deploy';

export const data = new SlashCommandBuilder()
  .setName('restart')
  .setDescription('Trigger a zero-downtime rolling restart of your server bot container')
  .setDefaultMemberPermissions(PermissionFlagsBits.ManageGuild);

async function doRestart(interaction: ChatInputCommandInteraction, dep: Deployment) {
  const result = await restartDeployment(dep.id);
  if (!result.success) {
    const errorEmbed = new EmbedBuilder()
      .setTitle('❌ Restart Failed')
      .setDescription(result.message)
      .setColor(0xed4245)
      .setTimestamp();

    if (interaction.replied || interaction.deferred) {
      await interaction.editReply({ embeds: [errorEmbed], components: [] });
    }
    return;
  }

  const successEmbed = new EmbedBuilder()
    .setTitle('🔄 Restart Triggered')
    .setDescription('A rolling restart of your bot container has been scheduled in the Kubernetes cluster. The bot will reconnect in ~15-30 seconds.')
    .setColor(0x57f287)
    .addFields(
      { name: 'Instance Label', value: dep.instance_label || 'Default', inline: true },
      { name: 'Bot Type', value: dep.bot_type.toUpperCase(), inline: true },
      { name: 'Deployment ID', value: dep.id, inline: true }
    )
    .setTimestamp();

  await interaction.editReply({ embeds: [successEmbed], components: [] });
}

export async function execute(interaction: ChatInputCommandInteraction) {
  if (!interaction.guildId) {
    await interaction.reply({ content: 'This command can only be used inside a server.', ephemeral: true });
    return;
  }

  await interaction.deferReply({ ephemeral: true });

  const deps = await getGuildDeployments(interaction.guildId);
  if (deps.length === 0) {
    await interaction.editReply({
      content: '❌ No active deployment was found for this server. If this is a shared bot, restarts are managed automatically.',
    });
    return;
  }

  // Single bot: restart immediately
  if (deps.length === 1) {
    await doRestart(interaction, deps[0]);
    return;
  }

  // Multi-bot: prompt with select menu
  const promptEmbed = new EmbedBuilder()
    .setTitle('🔄 Select Bot Instance to Restart')
    .setDescription(`This server has **${deps.length} active bot instances**. Choose which instance to perform a rolling restart on:`)
    .setColor(0x5865f2);

  const selectMenu = new StringSelectMenuBuilder()
    .setCustomId('select_restart_bot')
    .setPlaceholder('Select bot to restart...')
    .addOptions(
      deps.map((d, idx) =>
        new StringSelectMenuOptionBuilder()
          .setLabel(d.instance_label || `${d.bot_type} #${idx + 1}`)
          .setDescription(`Type: ${d.bot_type.toUpperCase()} | Status: ${d.status}`)
          .setValue(d.id)
      )
    );

  const row = new ActionRowBuilder<StringSelectMenuBuilder>().addComponents(selectMenu);
  const reply = await interaction.editReply({ embeds: [promptEmbed], components: [row] });

  const collector = reply.createMessageComponentCollector({
    componentType: ComponentType.StringSelect,
    time: 60_000,
  });

  collector.on('collect', async (i) => {
    if (i.user.id !== interaction.user.id) {
      await i.reply({ content: 'Only the user who ran /restart can use this menu.', ephemeral: true });
      return;
    }

    const selectedDepId = i.values[0];
    const selectedDep = deps.find((d) => d.id === selectedDepId);
    if (!selectedDep) return;

    await i.deferUpdate();
    await doRestart(interaction, selectedDep);
  });

  collector.on('end', async (_, reason) => {
    if (reason === 'time') {
      try {
        selectMenu.setDisabled(true);
        await interaction.editReply({ components: [row] });
      } catch {}
    }
  });
}
