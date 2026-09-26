import http from 'http';
import path from 'path';
import { ShardingManager } from 'discord.js';
import { config } from './config';

export function startSharder() {
  if (!config.discordToken) {
    console.warn('[Manager Bot Sharder] MANAGER_BOT_TOKEN not provided; starting fallback standalone bot.');
    const { startBotClient } = require('./bot');
    startBotClient(true);
    return;
  }

  // Determine whether running compiled JS or ts-node
  const isTs = __filename.endsWith('.ts');
  const botFile = isTs ? path.join(__dirname, 'bot.ts') : path.join(__dirname, 'bot.js');

  console.log(`[Manager Bot Sharder] Initializing ShardingManager targeting: ${botFile}`);

  const manager = new ShardingManager(botFile, {
    token: config.discordToken,
    totalShards: 'auto',
    execArgv: isTs ? ['-r', 'ts-node/register'] : [],
  });

  manager.on('shardCreate', (shard) => {
    console.log(`[Manager Bot Sharder] Launched Shard #${shard.id}`);

    shard.on('ready', () => {
      console.log(`[Manager Bot Sharder] Shard #${shard.id} connected to Discord Gateway`);
    });

    shard.on('disconnect', () => {
      console.warn(`[Manager Bot Sharder] Shard #${shard.id} disconnected`);
    });

    shard.on('reconnecting', () => {
      console.log(`[Manager Bot Sharder] Shard #${shard.id} reconnecting...`);
    });

    shard.on('error', (error) => {
      console.error(`[Manager Bot Sharder] Shard #${shard.id} encountered error:`, error);
    });
  });

  // Master HTTP Health Server aggregating metrics across all spawned shards
  const healthPort = Number(process.env.HEALTH_PORT || 8085);
  const healthServer = http.createServer(async (req, res) => {
    const url = req.url || '';
    if (
      url === '/health' ||
      url === '/livez' ||
      url === '/readyz' ||
      url === '/api/v1/manager-bot/health' ||
      url === '/api/v1/manager-bot/livez' ||
      url === '/api/v1/manager-bot/readyz'
    ) {
      let shardStats: Array<{ id: number; ready: boolean; ping: number; guilds: number }> = [];

      try {
        const evalResults = await manager.broadcastEval((client) => ({
          ready: client.isReady(),
          ping: client.ws?.ping ?? -1,
          guilds: client.guilds?.cache.size ?? 0,
        }));

        shardStats = evalResults.map((stat, idx) => ({
          id: idx,
          ...stat,
        }));
      } catch (err) {
        console.warn('[Manager Bot Sharder] broadcastEval error during health check:', err);
      }

      const totalGuilds = shardStats.reduce((sum, s) => sum + s.guilds, 0);
      const allReady = shardStats.length > 0 && shardStats.every((s) => s.ready);
      const avgPing =
        shardStats.length > 0
          ? Math.round(shardStats.reduce((sum, s) => sum + (s.ping > 0 ? s.ping : 0), 0) / shardStats.length)
          : -1;

      const statusCode = allReady || !config.discordToken ? 200 : 503;
      res.writeHead(statusCode, { 'Content-Type': 'application/json' });
      res.end(
        JSON.stringify({
          status: allReady ? 'healthy' : 'starting',
          service: 'manager-bot',
          mode: 'sharded',
          ready: allReady,
          total_shards: manager.shards.size,
          shards: shardStats,
          guilds_count: totalGuilds,
          avg_ping_ms: avgPing,
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
    console.log(`[Manager Bot Sharder] Master health aggregator listening on http://0.0.0.0:${healthPort}/health`);
  });

  manager.spawn().catch((err) => {
    console.error('[Manager Bot Sharder] Failed spawning shards:', err);
  });
}
