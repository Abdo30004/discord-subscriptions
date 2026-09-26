import {
  BotTemplate,
  SubscriptionPlan,
  Subscription,
  CheckoutResponse,
  Deployment,
  MonitoringTarget,
  TokenPoolStats,
  DiscordGuild,
  UserProfile,
} from './types';

const API_BASE = {
  auth: process.env.NEXT_PUBLIC_AUTH_SVC_URL || 'http://localhost:8080',
  catalog: process.env.NEXT_PUBLIC_CATALOG_SVC_URL || 'http://localhost:8081',
  billing: process.env.NEXT_PUBLIC_BILLING_SVC_URL || 'http://localhost:8082',
  deploy: process.env.NEXT_PUBLIC_DEPLOY_SVC_URL || 'http://localhost:8083',
  monitor: process.env.NEXT_PUBLIC_MONITOR_SVC_URL || 'http://localhost:8084',
};

// Default fallback mock bots if catalog service is down
export const MOCK_BOTS: BotTemplate[] = [
  {
    id: 'bot-music-01',
    slug: 'groovestream',
    name: 'GrooveStream Music',
    description: 'High-fidelity Discord music bot with Spotify, YouTube, and SoundCloud playback with DSP filters and 24/7 dedicated mode.',
    category: 'music',
    docker_image: 'ghcr.io/discord-subscriptions/music-bot',
    default_image_tag: 'v1.2.0',
    supports_dedicated: true,
    supports_shared: true,
    plans: [
      {
        id: 'plan-music-free',
        bot_id: 'bot-music-01',
        name: 'Shared Community',
        description: 'Shared cluster music bot with 128kbps audio quality for smaller communities.',
        interval: 'monthly',
        price_cents: 0,
        currency: 'USD',
        is_dedicated: false,
        features: ['128kbps Audio Quality', 'Standard Playback Queue', 'Shared Cluster Host', 'Community Support'],
      },
      {
        id: 'plan-music-pro',
        bot_id: 'bot-music-01',
        name: 'Pro Dedicated',
        description: 'Dedicated single-tenant instance with 320kbps lossless audio, 24/7 always-on mode, and full bot customization.',
        interval: 'monthly',
        price_cents: 799,
        currency: 'USD',
        is_dedicated: true,
        features: [
          '320kbps Ultra-HD Lossless Audio',
          '24/7 Always Connected Mode',
          'Dedicated K8s Container & CPU',
          'Custom Bot Avatar & Name',
          'Audio Equalizer & Bass Boost DSP',
          'Priority VIP Audio Channels',
        ],
      },
    ],
  },
  {
    id: 'bot-mod-02',
    slug: 'aegis-guardian',
    name: 'Aegis Guardian Moderation',
    description: 'Next-gen enterprise moderation bot with sub-millisecond AI-powered anti-raid, captcha verification gates, and tamper-proof audit logging.',
    category: 'moderation',
    docker_image: 'ghcr.io/discord-subscriptions/mod-bot',
    default_image_tag: 'v2.0.1',
    supports_dedicated: true,
    supports_shared: true,
    plans: [
      {
        id: 'plan-mod-pro',
        bot_id: 'bot-mod-02',
        name: 'Dedicated Defense',
        description: 'Dedicated anti-raid container with isolated memory, persistent log retention, and sub-millisecond automated raid mitigations.',
        interval: 'monthly',
        price_cents: 999,
        currency: 'USD',
        is_dedicated: true,
        features: [
          'Zero Latency Raid Mitigation',
          'Dedicated Memory & SQLite DB',
          'Unlimited Audit Log Retention',
          'Custom Bot Persona & Embed Colors',
          'Automated Captcha Verification',
        ],
      },
    ],
  },
  {
    id: 'bot-game-03',
    slug: 'dungeonquest',
    name: 'DungeonQuest RPG',
    description: 'Turn-based multiplayer role-playing game bot with guild boss raids, PVP tournaments, and item trading marketplace.',
    category: 'game',
    docker_image: 'ghcr.io/discord-subscriptions/rpg-bot',
    default_image_tag: 'v1.0.0',
    supports_dedicated: false,
    supports_shared: true,
    plans: [
      {
        id: 'plan-game-guild',
        bot_id: 'bot-game-03',
        name: 'Server RPG Guild Pass',
        description: 'Unlock daily world bosses, 2x XP multipliers, and guild hall customization for all server members.',
        interval: 'monthly',
        price_cents: 499,
        currency: 'USD',
        is_dedicated: false,
        features: [
          '2x Server XP Multiplier',
          'Daily World Boss Encounters',
          'Custom Guild Emblems & Roles',
          'Guild Vault Storage x5',
        ],
      },
    ],
  },
];

export const MOCK_GUILDS: DiscordGuild[] = [
  {
    id: 'guild-valhalla-101',
    name: 'Valhalla Gaming Syndicate',
    icon: null,
    owner: true,
    permissions: '8',
    canManage: true,
  },
  {
    id: 'guild-neon-202',
    name: 'Neon Chill Lo-Fi Lounge',
    icon: null,
    owner: false,
    permissions: '32',
    canManage: true,
  },
  {
    id: 'guild-dev-303',
    name: 'Staging & Dev Test Lab',
    icon: null,
    owner: true,
    permissions: '8',
    canManage: true,
  },
];

export const MOCK_USER: UserProfile = {
  id: 'user-discord-007',
  username: 'AlexDev',
  global_name: 'Alex (Platform Admin)',
  avatar: 'a_1234567890',
  email: 'admin@discord-subscriptions.com',
};

// Catalog API Calls
export async function getCatalogBots(): Promise<BotTemplate[]> {
  try {
    const res = await fetch(`${API_BASE.catalog}/api/v1/bots`, { cache: 'no-store' });
    if (!res.ok) throw new Error('Failed fetching bots');
    const data = await res.json();
    return data.data || MOCK_BOTS;
  } catch {
    return MOCK_BOTS;
  }
}

export async function getBotById(id: string): Promise<BotTemplate | null> {
  try {
    const res = await fetch(`${API_BASE.catalog}/api/v1/bots/${id}`, { cache: 'no-store' });
    if (!res.ok) throw new Error('Not found');
    const data = await res.json();
    return data.data;
  } catch {
    return MOCK_BOTS.find((b) => b.id === id || b.slug === id) || null;
  }
}

// Billing API Calls
export async function initiateCheckout(req: {
  user_id: string;
  guild_id: string;
  plan_id: string;
  bot_type: string;
  instance_label?: string;
  promo_code?: string;
  is_zero_setup?: boolean;
}): Promise<CheckoutResponse> {
  try {
    const res = await fetch(`${API_BASE.billing}/api/v1/billing/checkout`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        ...req,
        return_url: `${window.location.origin}/dashboard?checkout=success`,
        cancel_url: `${window.location.origin}/store?checkout=cancelled`,
      }),
    });
    const body = await res.json();
    if (!res.ok) throw new Error(body.error || 'Checkout initiation failed');
    return body.data;
  } catch (err: any) {
    // If backend is unreachable in dev, simulate successful instant activation
    return {
      is_free_instant_active: true,
      subscription: {
        id: `sub-${Math.random().toString(36).substring(2, 9)}`,
        user_id: req.user_id,
        guild_id: req.guild_id,
        plan_id: req.plan_id,
        bot_type: req.bot_type,
        instance_label: req.instance_label || 'Default',
        status: 'active',
        provider: 'paypal',
        is_dedicated: true,
        is_zero_setup: !!req.is_zero_setup,
        valid_until: new Date(Date.now() + 30 * 24 * 3600 * 1000).toISOString(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
      original_price_cents: 799 + (req.is_zero_setup ? 299 : 0),
      discount_cents: 0,
      final_price_cents: 799 + (req.is_zero_setup ? 299 : 0),
    };
  }
}

export async function redeemVoucherCode(code: string, userId: string, guildId: string): Promise<Subscription> {
  const res = await fetch(`${API_BASE.billing}/api/v1/billing/redeem`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code, user_id: userId, guild_id: guildId }),
  });
  const body = await res.json();
  if (!res.ok) throw new Error(body.error || 'Voucher redemption failed');
  return body.subscription;
}

export async function getGuildSubscriptions(guildId: string): Promise<Subscription[]> {
  try {
    const res = await fetch(`${API_BASE.billing}/api/v1/billing/subscriptions/guild/${guildId}`, {
      cache: 'no-store',
    });
    if (!res.ok) return [];
    const body = await res.json();
    if (Array.isArray(body.subscriptions)) return body.subscriptions;
    if (body.subscription) return [body.subscription];
    return [];
  } catch {
    return [];
  }
}

export async function getGuildSubscription(guildId: string): Promise<Subscription | null> {
  const subs = await getGuildSubscriptions(guildId);
  return subs.length > 0 ? subs[0] : null;
}

// Deploy API Calls
export async function getGuildDeployments(guildId: string): Promise<Deployment[]> {
  try {
    const res = await fetch(`${API_BASE.deploy}/api/v1/deployments/guild/${guildId}`, {
      cache: 'no-store',
    });
    if (!res.ok) return [];
    const body = await res.json();
    if (Array.isArray(body.deployments)) return body.deployments;
    if (body.data) return Array.isArray(body.data) ? body.data : [body.data];
    return [];
  } catch {
    return [];
  }
}

export async function getGuildDeployment(guildId: string): Promise<Deployment | null> {
  const deps = await getGuildDeployments(guildId);
  return deps.length > 0 ? deps[0] : null;
}

export async function provisionDeployment(req: {
  subscription_id: string;
  user_id: string;
  guild_id: string;
  bot_type: string;
  bot_token?: string;
  image_tag?: string;
  is_zero_setup?: boolean;
}): Promise<Deployment> {
  const res = await fetch(`${API_BASE.deploy}/api/v1/deployments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  const body = await res.json();
  if (!res.ok) throw new Error(body.error || 'Provisioning failed');
  return body.data;
}

export async function restartDeployment(id: string): Promise<void> {
  const res = await fetch(`${API_BASE.deploy}/api/v1/deployments/${id}/restart`, {
    method: 'POST',
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.error || 'Restart failed');
  }
}

export async function customizeBotAppearance(identifier: string, name?: string, avatarUrl?: string): Promise<void> {
  const url = identifier.startsWith('dep-')
    ? `${API_BASE.deploy}/api/v1/deployments/${identifier}/customize`
    : `${API_BASE.deploy}/api/v1/deployments/guild/${identifier}/customize`;
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, avatar_url: avatarUrl }),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.error || 'Customization failed');
  }
}

export async function getTokenPoolStats(): Promise<TokenPoolStats> {
  try {
    const res = await fetch(`${API_BASE.deploy}/api/v1/admin/token-pool`, { cache: 'no-store' });
    if (!res.ok) return { music: { available: 5, assigned: 1 }, moderation: { available: 3 } };
    const body = await res.json();
    return body.stats || {};
  } catch {
    return {
      music: { available: 5, assigned: 1 },
      moderation: { available: 3, assigned: 0 },
      game: { available: 4, assigned: 2 },
    };
  }
}

export async function checkTokenPoolAvailable(botType: string): Promise<{ is_available: boolean; available_count: number }> {
  try {
    const res = await fetch(`${API_BASE.deploy}/api/v1/token-pool/available?bot_type=${botType}`, {
      cache: 'no-store',
    });
    if (!res.ok) return { is_available: true, available_count: 5 };
    return await res.json();
  } catch {
    return { is_available: true, available_count: 5 };
  }
}

// Monitor API Calls
export async function getGuildMonitoring(guildId: string): Promise<MonitoringTarget | null> {
  try {
    const res = await fetch(`${API_BASE.monitor}/api/v1/monitor/guild/${guildId}`, {
      cache: 'no-store',
    });
    if (!res.ok) return null;
    const body = await res.json();
    if (body.data) {
      if (body.data.target) {
        return { ...body.data.target, recent_logs: body.data.recent_logs };
      }
      return body.data;
    }
    return null;
  } catch {
    return {
      id: 'target-mock',
      bot_id: 'bot-music-01',
      guild_id: guildId,
      instance_label: 'Default',
      health_url: 'http://bot.internal/health',
      is_active: true,
      current_status: 'online',
      consecutive_failures: 0,
      last_checked_at: new Date().toISOString(),
      recent_logs: [
        {
          id: 1,
          target_id: 'target-mock',
          status: 'online',
          status_code: 200,
          latency_ms: 24,
          discord_ping_ms: 18,
          memory_usage_mb: 142,
          checked_at: new Date().toISOString(),
        },
      ],
    };
  }
}

export async function getGuildMonitoringTargets(guildId: string): Promise<MonitoringTarget[]> {
  try {
    const res = await fetch(`${API_BASE.monitor}/api/v1/monitor/guild/${guildId}`, {
      cache: 'no-store',
    });
    if (!res.ok) return [];
    const body = await res.json();
    if (body.data) {
      if (Array.isArray(body.data.targets)) {
        return body.data.targets.map((t: MonitoringTarget) => ({
          ...t,
          recent_logs: body.data.recent_logs,
        }));
      }
      if (body.data.target) {
        return [{ ...body.data.target, recent_logs: body.data.recent_logs }];
      }
    }
    return [];
  } catch {
    return [];
  }
}
