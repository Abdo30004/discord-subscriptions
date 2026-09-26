import http from 'http';
import { Client, GatewayIntentBits, Events } from 'discord.js';
import { config } from './config';
import { handleReady } from './events/ready';
import { handleInteraction } from './events/interactionCreate';
import { deployCommands } from './deploy-commands';

const client = new Client({
  intents: [GatewayIntentBits.Guilds],
});

client.once(Events.ClientReady, (readyClient) => {
  handleReady(readyClient);
  // Auto-deploy commands on ready if configured
  deployCommands().catch(console.error);
});

client.on(Events.InteractionCreate, (interaction) => {
  handleInteraction(interaction);
});

if (!config.discordToken) {
  console.warn('[Manager Bot] MANAGER_BOT_TOKEN is not configured in .env. Bot is running in offline standby mode.');
} else {
  client.login(config.discordToken).catch((err) => {
    console.error('[Manager Bot] Failed to login to Discord Gateway:', err);
  });
}

// Lightweight HTTP Health Check Server
const healthPort = Number(process.env.HEALTH_PORT || 8085);
const healthServer = http.createServer((req, res) => {
  const url = req.url || '';
  if (
    url === '/health' ||
    url === '/livez' ||
    url === '/readyz' ||
    url === '/api/v1/manager-bot/health' ||
    url === '/api/v1/manager-bot/livez' ||
    url === '/api/v1/manager-bot/readyz'
  ) {
    const isReady = client.isReady();
    // In dev without token, standby is treated as 200 OK so container stays up
    const statusCode = isReady || !config.discordToken ? 200 : 503;
    res.writeHead(statusCode, { 'Content-Type': 'application/json' });
    res.end(
      JSON.stringify({
        status: isReady ? 'healthy' : (!config.discordToken ? 'standby' : 'starting'),
        service: 'manager-bot',
        ready: isReady,
        network: 'platform-net',
        traefik_gateway: config.gatewayUrl,
        ping_ms: client.ws?.ping ?? -1,
        guilds_count: client.guilds?.cache.size ?? 0,
        uptime_seconds: Math.floor(process.uptime()),
        timestamp: new Date().toISOString(),
      })
    );
    return;
  }
  res.writeHead(404, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify({ error: 'not found' }));
});

healthServer.listen(healthPort, () => {
  console.log(`[Manager Bot] Health check probe listening on http://0.0.0.0:${healthPort}/health`);
});
