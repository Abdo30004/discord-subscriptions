import { startSharder } from './sharder';
import { startBotClient } from './bot';

const enableSharding = process.env.ENABLE_SHARDING === 'true';

if (enableSharding) {
  console.log('[Manager Bot] ENABLE_SHARDING=true; launching cluster via ShardingManager...');
  startSharder();
} else {
  console.log('[Manager Bot] Starting bot in single-process standalone mode (ENABLE_SHARDING=false)...');
  startBotClient(true);
}
