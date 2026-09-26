import { config } from '../config';

export interface HealthLog {
  status: string;
  status_code: number;
  latency_ms: number;
  discord_ping_ms: number;
  memory_usage_mb: number;
  checked_at: string;
}

export interface TargetInfo {
  id?: string;
  bot_id: string;
  guild_id: string;
  instance_label?: string;
  current_status: string;
  last_checked_at?: string;
}

export interface GuildStatusResponse {
  target?: TargetInfo;
  targets?: TargetInfo[];
  recent_logs: HealthLog[];
}

export async function getGuildLiveStatus(guildId: string): Promise<GuildStatusResponse | null> {
  try {
    const res = await fetch(`${config.monitorSvcUrl}/api/v1/monitor/guild/${guildId}`);
    if (!res.ok) return null;
    const body = await res.json();
    return body.data || null;
  } catch {
    return null;
  }
}
