import {
  BotTemplate,
  SubscriptionPlan,
  Subscription,
  CheckoutResponse,
  Deployment,
  MonitoringTarget,
  HealthLog,
  TokenPoolStats,
  DiscordGuild,
  UserProfile,
  AuthSession,
  CheckoutRequest,
  ProvisionRequest,
  AdminGrantRequest,
  CreatePromoRequest,
  CreateVoucherRequest,
  PromoCode,
  VoucherCode,
} from './types';

// Traefik reverse proxy handles path prefix routing to all upstream microservices.
// In the browser, all requests default to relative URLs (''), going directly to Traefik on port 80.
const API_BASE = {
  auth: process.env.NEXT_PUBLIC_AUTH_SVC_URL ?? '',
  catalog: process.env.NEXT_PUBLIC_CATALOG_SVC_URL ?? '',
  billing: process.env.NEXT_PUBLIC_BILLING_SVC_URL ?? '',
  deploy: process.env.NEXT_PUBLIC_DEPLOY_SVC_URL ?? '',
  monitor: process.env.NEXT_PUBLIC_MONITOR_SVC_URL ?? '',
};

// ============================================================================
// Auth Service API Calls
// ============================================================================

/**
 * Retrieves the Discord OAuth2 authorization URL from auth-svc.
 */
export async function getDiscordOAuthUrl(redirectUri?: string): Promise<string> {
  const uri = redirectUri || `${window.location.origin}/auth/callback`;
  const res = await fetch(`${API_BASE.auth}/api/v1/auth/discord/url?redirect_uri=${encodeURIComponent(uri)}`, {
    cache: 'no-store',
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || 'Failed to generate Discord login URL');
  }
  const body = await res.json();
  return body.url;
}

/**
 * Exchanges a Discord OAuth2 code for an authenticated session and JWT.
 */
export async function authenticateWithDiscord(code: string, redirectUri?: string): Promise<AuthSession> {
  const uri = redirectUri || `${window.location.origin}/auth/callback`;
  const res = await fetch(`${API_BASE.auth}/api/v1/auth/discord/callback`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code, redirect_uri: uri }),
  });
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || 'Discord authentication failed');
  }
  return body.data;
}

/**
 * Generates an instant development session for local testing.
 */
export async function devLogin(userId?: string, username?: string): Promise<AuthSession> {
  const res = await fetch(`${API_BASE.auth}/api/v1/auth/dev-login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      user_id: userId || '123456789012345678',
      username: username || 'DevAdmin',
    }),
  });
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || 'Development login failed');
  }
  return body.data;
}

/**
 * Retrieves the authenticated user profile using their JWT session.
 */
export async function getCurrentUser(token: string): Promise<UserProfile> {
  const res = await fetch(`${API_BASE.auth}/api/v1/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  });
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || 'Failed to retrieve user profile');
  }
  return body.data;
}

/**
 * Retrieves all Discord servers where the authenticated user has management permissions.
 */
export async function getUserGuilds(token: string): Promise<DiscordGuild[]> {
  const res = await fetch(`${API_BASE.auth}/api/v1/auth/guilds`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  });
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || 'Failed to retrieve user servers');
  }
  return body.data || [];
}

// ============================================================================
// Catalog Service API Calls
// ============================================================================

/**
 * Fetches all available bot templates and their pricing plans from catalog-svc.
 */
export async function getCatalogBots(): Promise<BotTemplate[]> {
  const res = await fetch(`${API_BASE.catalog}/api/v1/bots`, { cache: 'no-store' });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || 'Failed to retrieve bot catalog');
  }
  const body = await res.json();
  return body.data || [];
}

/**
 * Fetches single bot template details by ID or slug.
 */
export async function getBotById(idOrSlug: string): Promise<BotTemplate> {
  const res = await fetch(`${API_BASE.catalog}/api/v1/bots/${idOrSlug}`, { cache: 'no-store' });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Bot "${idOrSlug}" not found in catalog`);
  }
  const body = await res.json();
  return body.data;
}

/**
 * Fetches subscription plan details by ID.
 */
export async function getPlanById(id: string): Promise<SubscriptionPlan> {
  const res = await fetch(`${API_BASE.catalog}/api/v1/plans/${id}`, { cache: 'no-store' });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Plan "${id}" not found`);
  }
  const body = await res.json();
  return body.data;
}

// ============================================================================
// Billing Service API Calls
// ============================================================================

/**
 * Initiates checkout for a subscription plan or turnkey zero-setup delivery.
 */
export async function initiateCheckout(req: CheckoutRequest): Promise<CheckoutResponse> {
  const payload = {
    ...req,
    return_url: req.return_url || `${window.location.origin}/dashboard?checkout=success`,
    cancel_url: req.cancel_url || `${window.location.origin}/store?checkout=cancelled`,
  };

  const res = await fetch(`${API_BASE.billing}/api/v1/billing/checkout`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || 'Checkout initiation failed');
  }
  return body.data;
}

/**
 * Redeems a promotional gift voucher code for a guild subscription.
 */
export async function redeemVoucherCode(code: string, userId: string, guildId: string): Promise<Subscription> {
  const res = await fetch(`${API_BASE.billing}/api/v1/billing/redeem`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code, user_id: userId, guild_id: guildId }),
  });
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || 'Voucher redemption failed');
  }
  return body.subscription;
}

/**
 * Retrieves all subscriptions for a specific Discord guild.
 */
export async function getGuildSubscriptions(guildId: string): Promise<Subscription[]> {
  const res = await fetch(`${API_BASE.billing}/api/v1/billing/subscriptions/guild/${guildId}`, {
    cache: 'no-store',
  });
  if (!res.ok) {
    return [];
  }
  const body = await res.json();
  if (Array.isArray(body.subscriptions)) return body.subscriptions;
  if (body.subscription) return [body.subscription];
  return [];
}

/**
 * Creates a promotional discount code (admin only).
 */
export async function createPromoCode(req: CreatePromoRequest): Promise<PromoCode> {
  const res = await fetch(`${API_BASE.billing}/api/v1/billing/admin/promo`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || 'Failed to create promo code');
  }
  return body.promo;
}

/**
 * Creates a gift voucher code (admin only).
 */
export async function createVoucherCode(req: CreateVoucherRequest): Promise<VoucherCode> {
  const res = await fetch(`${API_BASE.billing}/api/v1/billing/admin/voucher`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || 'Failed to create voucher code');
  }
  return body.voucher;
}

/**
 * Grants an instant administrator subscription without payment.
 */
export async function adminGrantSubscription(req: AdminGrantRequest): Promise<Subscription> {
  const res = await fetch(`${API_BASE.billing}/api/v1/billing/admin/grant`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || 'Failed to grant subscription');
  }
  return body.subscription;
}

// ============================================================================
// Deploy Service API Calls
// ============================================================================

/**
 * Retrieves all active bot deployments for a specific Discord guild.
 */
export async function getGuildDeployments(guildId: string): Promise<Deployment[]> {
  const res = await fetch(`${API_BASE.deploy}/api/v1/deployments/guild/${guildId}`, {
    cache: 'no-store',
  });
  if (!res.ok) {
    return [];
  }
  const body = await res.json();
  if (Array.isArray(body.deployments)) return body.deployments;
  if (body.data) return Array.isArray(body.data) ? body.data : [body.data];
  return [];
}

/**
 * Retrieves a single deployment by its unique ID.
 */
export async function getDeploymentById(id: string): Promise<Deployment> {
  const res = await fetch(`${API_BASE.deploy}/api/v1/deployments/${id}`, { cache: 'no-store' });
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || 'Failed to retrieve deployment details');
  }
  return body.data;
}

/**
 * Provisions a bot pod in Kubernetes or allocates from the pre-warmed turnkey token pool.
 */
export async function provisionDeployment(req: ProvisionRequest): Promise<Deployment> {
  const res = await fetch(`${API_BASE.deploy}/api/v1/deployments`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  });
  const body = await res.json();
  if (!res.ok) {
    throw new Error(body.error || 'Bot provisioning failed');
  }
  return body.data;
}

/**
 * Triggers a rolling pod restart in Kubernetes.
 */
export async function restartDeployment(id: string): Promise<void> {
  const res = await fetch(`${API_BASE.deploy}/api/v1/deployments/${id}/restart`, {
    method: 'POST',
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || 'Bot restart request failed');
  }
}

/**
 * Customizes the bot's Discord username and avatar persona in the target guild.
 */
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
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || 'Persona customization failed');
  }
}

/**
 * Retrieves token pool inventory statistics from deploy-svc.
 */
export async function getTokenPoolStats(): Promise<TokenPoolStats> {
  const res = await fetch(`${API_BASE.deploy}/api/v1/admin/token-pool`, { cache: 'no-store' });
  if (!res.ok) {
    return {};
  }
  const body = await res.json();
  return body.stats || {};
}

/**
 * Checks whether pre-warmed turnkey tokens are available for instant 0-setup deployment.
 */
export async function checkTokenPoolAvailable(botType: string): Promise<{ is_available: boolean; available_count: number }> {
  const res = await fetch(`${API_BASE.deploy}/api/v1/token-pool/available?bot_type=${botType}`, {
    cache: 'no-store',
  });
  if (!res.ok) {
    return { is_available: false, available_count: 0 };
  }
  return await res.json();
}

/**
 * Adds fresh pre-warmed tokens into HashiCorp Vault and deploy_db (admin only).
 */
export async function addPoolTokens(botType: string, tokens: Array<{ token: string; client_id: string }>): Promise<void> {
  const res = await fetch(`${API_BASE.deploy}/api/v1/admin/token-pool`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bot_type: botType, tokens }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || 'Failed to add tokens to pool');
  }
}

// ============================================================================
// Monitor Service API Calls
// ============================================================================

/**
 * Retrieves health monitoring targets and real-time telemetry for a guild.
 */
export async function getGuildMonitoringTargets(guildId: string): Promise<MonitoringTarget[]> {
  const res = await fetch(`${API_BASE.monitor}/api/v1/monitor/guild/${guildId}`, {
    cache: 'no-store',
  });
  if (!res.ok) {
    return [];
  }
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
}

/**
 * Retrieves detailed health logs and active targets for a guild.
 */
export async function getGuildStatus(guildId: string): Promise<{
  targets: MonitoringTarget[];
  target?: MonitoringTarget;
  recent_logs: HealthLog[];
}> {
  const res = await fetch(`${API_BASE.monitor}/api/v1/monitor/guild/${guildId}`, {
    cache: 'no-store',
  });
  if (!res.ok) {
    return { targets: [], recent_logs: [] };
  }
  const body = await res.json();
  return {
    targets: body.data?.targets || (body.data?.target ? [body.data.target] : []),
    target: body.data?.target,
    recent_logs: body.data?.recent_logs || [],
  };
}
