import { config } from '../config';

export interface Subscription {
  id: string;
  user_id: string;
  guild_id: string;
  plan_id: string;
  bot_type: string;
  instance_label?: string;
  status: string;
  provider: string;
  is_dedicated: boolean;
  is_zero_setup?: boolean;
  valid_until: string;
}

export async function redeemVoucherCode(code: string, userId: string, guildId: string): Promise<{ success: boolean; message: string; subscription?: Subscription }> {
  try {
    const res = await fetch(`${config.billingSvcUrl}/api/v1/billing/redeem`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        code,
        user_id: userId,
        guild_id: guildId,
      }),
    });

    const body = await res.json();
    if (!res.ok) {
      return { success: false, message: body.error || 'Failed to redeem voucher' };
    }

    return { success: true, message: body.message, subscription: body.subscription };
  } catch (err: any) {
    return { success: false, message: err.message || 'Internal connection error' };
  }
}

export async function getGuildSubscriptions(guildId: string): Promise<Subscription[]> {
  try {
    const res = await fetch(`${config.billingSvcUrl}/api/v1/billing/subscriptions/guild/${guildId}`);
    if (!res.ok) return [];
    const body = await res.json();
    if (Array.isArray(body.subscriptions)) {
      return body.subscriptions;
    }
    if (body.subscription) {
      return [body.subscription];
    }
    return [];
  } catch {
    return [];
  }
}

export async function getGuildSubscription(guildId: string): Promise<Subscription | null> {
  const subs = await getGuildSubscriptions(guildId);
  return subs.length > 0 ? subs[0] : null;
}
