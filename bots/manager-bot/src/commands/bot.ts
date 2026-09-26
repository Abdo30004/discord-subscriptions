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
import { customizeBot, getGuildDeployments, Deployment } from '../api/deploy';

export const data = new SlashCommandBuilder()
  .setName('bot')
  .setDescription('Manage and customize your server bot appearance')
  .setDefaultMemberPermissions(PermissionFlagsBits.ManageGuild)
  .addSubcommand((sub) =>
    sub
      .setName('name')
      .setDescription('Change bot username (Discord limits this to 2 changes per hour)')
      .addStringOption((opt) =>
        opt
          .setName('value')
          .setDescription('The new username for the bot')
          .setRequired(true)
          .setMinLength(2)
          .setMaxLength(32)
      )
  )
  .addSubcommand((sub) =>
    sub
      .setName('avatar')
      .setDescription('Change bot avatar image')
      .addStringOption((opt) =>
        opt
          .setName('url')
          .setDescription('Direct URL to image (PNG or JPG format)')
          .setRequired(true)
      )
  );

async function applyCustomization(
  interaction: ChatInputCommandInteraction,
  dep: Deployment,
  subcommand: string,
  options: { name?: string; avatarUrl?: string }
) {
  const result = await customizeBot(dep.id, options);

  if (!result.success) {
    const errEmbed = new EmbedBuilder()
      .setTitle(`❌ ${subcommand === 'name' ? 'Username' : 'Avatar'} Update Failed`)
      .setDescription(result.message)
      .setColor(0xed4245)
      .setTimestamp();
    await interaction.editReply({ embeds: [errEmbed], components: [] });
    return;
  }

  if (subcommand === 'name') {
    const embed = new EmbedBuilder()
      .setTitle('✨ Bot Username Updated')
      .setDescription(`Bot username for **${dep.instance_label || dep.bot_type}** has been successfully updated to **${options.name}**!`)
      .setColor(0x57f287)
      .addFields(
        { name: 'Instance Label', value: dep.instance_label || 'Default', inline: true },
        { name: 'Bot Type', value: dep.bot_type.toUpperCase(), inline: true },
        { name: 'Note', value: 'Discord enforces a maximum of 2 name changes per hour.', inline: true }
      )
      .setTimestamp();

    await interaction.editReply({ embeds: [embed], components: [] });
  } else {
    const embed = new EmbedBuilder()
      .setTitle('🎨 Bot Avatar Updated')
      .setDescription(`Bot avatar picture for **${dep.instance_label || dep.bot_type}** has been updated!`)
      .setThumbnail(options.avatarUrl!)
      .setColor(0x57f287)
      .addFields(
        { name: 'Instance Label', value: dep.instance_label || 'Default', inline: true },
        { name: 'Bot Type', value: dep.bot_type.toUpperCase(), inline: true }
      )
      .setTimestamp();

    await interaction.editReply({ embeds: [embed], components: [] });
  }
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
      content: '❌ No active dedicated or managed bot deployment was found for this server.',
    });
    return;
  }

  const subcommand = interaction.options.getSubcommand();
  const options = {
    name: subcommand === 'name' ? interaction.options.getString('value', true) : undefined,
    avatarUrl: subcommand === 'avatar' ? interaction.options.getString('url', true) : undefined,
  };

  // Single bot: apply directly
  if (deps.length === 1) {
    await applyCustomization(interaction, deps[0], subcommand, options);
    return;
  }

  // Multi-bot: prompt with select menu
  const promptEmbed = new EmbedBuilder()
    .setTitle('🎨 Select Bot Instance to Customize')
    .setDescription(`This server has **${deps.length} active bot instances**. Choose which instance to customize:`)
    .setColor(0x5865f2);

  const selectMenu = new StringSelectMenuBuilder()
    .setCustomId('select_customize_bot')
    .setPlaceholder('Select bot to customize...')
    .addOptions(
      deps.map((d, idx) =>
        new StringSelectMenuOptionBuilder()
          .setLabel(d.instance_label || `${d.bot_type} #${idx + 1}`)
          .setDescription(`Type: ${d.bot_type.toUpperCase()} | ID: ${d.id}`)
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
      await i.reply({ content: 'Only the user who ran /bot can use this menu.', ephemeral: true });
      return;
    }

    const selectedDepId = i.values[0];
    const selectedDep = deps.find((d) => d.id === selectedDepId);
    if (!selectedDep) return;

    await i.deferUpdate();
    await applyCustomization(interaction, selectedDep, subcommand, options);
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
