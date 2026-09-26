// Catalog Service Types
export interface BotTemplate {
  id: string;
  slug: string;
  name: string;
  description: string;
  category: 'music' | 'moderation' | 'game' | 'utility';
  docker_image: string;
  default_image_tag: string;
  supports_dedicated: boolean;
  supports_shared: boolean;
  plans?: SubscriptionPlan[];
}

export interface SubscriptionPlan {
  id: string;
  bot_id: string;
  name: string;
  description: string;
  interval: string;
  price_cents: number;
  currency: string;
  is_dedicated: boolean;
  features: string[];
}

// Billing Service Types
export type SubscriptionStatus = 'active' | 'cancelled' | 'expired' | 'pending';
export type PaymentProvider = 'paypal' | 'gift_code' | 'admin_grant';

export interface Subscription {
  id: string;
  user_id: string;
  guild_id: string;
  plan_id: string;
  bot_type: string;
  instance_label?: string;
  status: SubscriptionStatus;
  provider: PaymentProvider;
  external_sub_id?: string;
  is_dedicated: boolean;
  is_zero_setup: boolean;
  valid_until: string;
  created_at: string;
  updated_at: string;
}

export interface PromoCode {
  id: string;
  code: string;
  discount_type: 'percentage' | 'fixed';
  discount_value: number;
  max_uses: number;
  current_uses: number;
  is_active: boolean;
  expires_at?: string;
}

export interface VoucherCode {
  id: string;
  code: string;
  plan_id: string;
  bot_type: string;
  duration_days: number;
  is_dedicated: boolean;
  is_redeemed: boolean;
  redeemed_by_user_id?: string;
  redeemed_by_guild_id?: string;
}

export interface CheckoutResponse {
  is_free_instant_active: boolean;
  subscription?: Subscription;
  approval_url?: string;
  original_price_cents: number;
  discount_cents: number;
  final_price_cents: number;
}

// Deployment Service Types
export type DeploymentStatus = 'pending' | 'deploying' | 'running' | 'failed' | 'stopped';

export interface Deployment {
  id: string;
  subscription_id: string;
  user_id: string;
  guild_id: string;
  bot_type: string;
  instance_label?: string;
  k8s_namespace: string;
  k8s_deployment_name: string;
  image_name: string;
  image_tag: string;
  status: DeploymentStatus;
  is_zero_setup: boolean;
  client_id?: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface TokenPoolStats {
  [botType: string]: {
    [status: string]: number;
  };
}

// Monitor Service Types
export interface MonitoringTarget {
  id: string;
  bot_id: string;
  guild_id: string;
  instance_label?: string;
  health_url: string;
  is_active: boolean;
  current_status: 'online' | 'degraded' | 'offline';
  consecutive_failures: number;
  last_checked_at?: string;
  recent_logs?: HealthLog[];
}

export interface HealthLog {
  id: number;
  target_id: string;
  status: string;
  status_code: number;
  latency_ms: number;
  discord_ping_ms: number;
  memory_usage_mb: number;
  checked_at: string;
  error_message?: string;
}

// Auth & Guild Types
export interface DiscordGuild {
  id: string;
  name: string;
  icon: string | null;
  owner: boolean;
  permissions: string;
  canManage: boolean;
}

export interface UserProfile {
  id: string;
  username: string;
  global_name?: string;
  avatar?: string;
  email?: string;
}

export interface AuthSession {
  token: string;
  expires_at: string;
  user: UserProfile;
}

export interface CheckoutRequest {
  user_id: string;
  guild_id: string;
  plan_id: string;
  bot_type: string;
  instance_label?: string;
  promo_code?: string;
  is_zero_setup?: boolean;
  return_url?: string;
  cancel_url?: string;
}

export interface ProvisionRequest {
  subscription_id: string;
  user_id: string;
  guild_id: string;
  bot_type: string;
  instance_label?: string;
  bot_token?: string;
  image_tag?: string;
  is_zero_setup?: boolean;
}

export interface AdminGrantRequest {
  user_id: string;
  guild_id: string;
  bot_type: string;
  plan_id: string;
  instance_label?: string;
  duration_days: number;
  is_dedicated: boolean;
  is_zero_setup: boolean;
}

export interface CreatePromoRequest {
  code: string;
  discount_type: 'percentage' | 'fixed';
  discount_value: number;
  max_uses: number;
  duration_days?: number;
}

export interface CreateVoucherRequest {
  code?: string;
  plan_id: string;
  bot_type: string;
  duration_days: number;
  is_dedicated: boolean;
}
