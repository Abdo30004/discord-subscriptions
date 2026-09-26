import { Collection } from 'discord.js';
import * as storeCommand from './store';
import * as redeemCommand from './redeem';
import * as statusCommand from './status';
import * as restartCommand from './restart';
import * as botCommand from './bot';
import * as helpCommand from './help';

export const commands = new Collection<string, any>();

const commandList = [storeCommand, redeemCommand, statusCommand, restartCommand, botCommand, helpCommand];

for (const cmd of commandList) {
  commands.set(cmd.data.name, cmd);
}

export const commandData = commandList.map(cmd => cmd.data.toJSON());
