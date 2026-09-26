'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import {
  ShieldCheck,
  ShieldAlert,
  Zap,
  Tag,
  Gift,
  Plus,
  CheckCircle2,
  AlertCircle,
  Database,
  Layers,
  Key,
  Loader2,
  UserCheck,
} from 'lucide-react';
import {
  getTokenPoolStats,
  addPoolTokens,
  createPromoCode,
  createVoucherCode,
  adminGrantSubscription,
} from '@/lib/api';
import { TokenPoolStats } from '@/lib/types';
import { useAuth } from '@/contexts/AuthContext';

export default function AdminPage() {
  const router = useRouter();
  const { user, isAdmin, isLoading: authLoading } = useAuth();

  const [stats, setStats] = useState<TokenPoolStats>({});
  const [activeTab, setActiveTab] = useState<'pool' | 'promo' | 'voucher' | 'grant'>('pool');
  const [loading, setLoading] = useState(false);
  const [fetchingStats, setFetchingStats] = useState(true);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Add Token State
  const [newBotType, setNewBotType] = useState('music');
  const [newClientId, setNewClientId] = useState('');
  const [newToken, setNewToken] = useState('');

  // Promo Code State
  const [promoCode, setPromoCode] = useState('');
  const [promoDiscountType, setPromoDiscountType] = useState<'percentage' | 'fixed'>('percentage');
  const [promoDiscountValue, setPromoDiscountValue] = useState(50);
  const [promoMaxUses, setPromoMaxUses] = useState(100);

  // Voucher Code State
  const [voucherCode, setVoucherCode] = useState('');
  const [voucherPlanId, setVoucherPlanId] = useState('plan-music-pro');
  const [voucherBotType, setVoucherBotType] = useState('music');
  const [voucherDays, setVoucherDays] = useState(30);

  // Admin Grant State
  const [grantUserId, setGrantUserId] = useState('');
  const [grantGuildId, setGrantGuildId] = useState('');
  const [grantBotType, setGrantBotType] = useState('music');
  const [grantPlanId, setGrantPlanId] = useState('plan-music-pro');
  const [grantDurationDays, setGrantDurationDays] = useState(30);

  const fetchStats = async () => {
    setFetchingStats(true);
    try {
      const res = await getTokenPoolStats();
      setStats(res);
    } catch (err: any) {
      console.error('Failed to load token pool stats:', err);
    } finally {
      setFetchingStats(false);
    }
  };

  useEffect(() => {
    if (!authLoading && (!user || !isAdmin)) {
      router.replace('/?auth=admin_required');
    }
  }, [user, isAdmin, authLoading, router]);

  useEffect(() => {
    if (user && isAdmin) {
      fetchStats();
    }
  }, [user, isAdmin]);

  const handleAddToken = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newClientId.trim() || !newToken.trim()) return;

    setLoading(true);
    setError(null);
    try {
      await addPoolTokens(newBotType, [
        {
          token: newToken.trim(),
          client_id: newClientId.trim(),
        },
      ]);
      setMessage(`Pre-warmed token for ${newBotType.toUpperCase()} safely stored in HashiCorp Vault!`);
      setNewClientId('');
      setNewToken('');
      await fetchStats();
      setTimeout(() => setMessage(null), 4000);
    } catch (err: any) {
      setError(err.message || 'Failed adding token to pool');
    } finally {
      setLoading(false);
    }
  };

  const handleCreatePromo = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!promoCode.trim()) return;

    setLoading(true);
    setError(null);
    try {
      await createPromoCode({
        code: promoCode.trim().toUpperCase(),
        discount_type: promoDiscountType,
        discount_value: Number(promoDiscountValue),
        max_uses: Number(promoMaxUses),
      });
      setMessage(`Promo code "${promoCode.toUpperCase()}" created and active in billing-svc!`);
      setPromoCode('');
      setTimeout(() => setMessage(null), 4000);
    } catch (err: any) {
      setError(err.message || 'Failed creating promo code');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateVoucher = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      const res = await createVoucherCode({
        code: voucherCode.trim() || undefined,
        plan_id: voucherPlanId,
        bot_type: voucherBotType,
        duration_days: Number(voucherDays),
        is_dedicated: true,
      });
      setMessage(`Gift voucher "${res.code}" created successfully!`);
      setVoucherCode('');
      setTimeout(() => setMessage(null), 5000);
    } catch (err: any) {
      setError(err.message || 'Failed creating gift voucher');
    } finally {
      setLoading(false);
    }
  };

  const handleAdminGrant = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!grantUserId.trim() || !grantGuildId.trim()) return;

    setLoading(true);
    setError(null);
    try {
      await adminGrantSubscription({
        user_id: grantUserId.trim(),
        guild_id: grantGuildId.trim(),
        bot_type: grantBotType,
        plan_id: grantPlanId,
        duration_days: Number(grantDurationDays),
        is_dedicated: true,
        is_zero_setup: true,
      });
      setMessage(`Admin subscription granted for guild ${grantGuildId.trim()}! Pod provisioning queued.`);
      setGrantUserId('');
      setGrantGuildId('');
      setTimeout(() => setMessage(null), 5000);
    } catch (err: any) {
      setError(err.message || 'Failed granting subscription');
    } finally {
      setLoading(false);
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-[70vh] flex flex-col items-center justify-center gap-4 text-center px-4">
        <Loader2 className="w-10 h-10 animate-spin text-blurple" />
        <p className="text-slate-400 text-sm">Verifying administrator authorization...</p>
      </div>
    );
  }

  if (!user || !isAdmin) {
    return (
      <div className="min-h-[70vh] flex flex-col items-center justify-center gap-4 text-center px-4">
        <ShieldAlert className="w-12 h-12 text-rose-500" />
        <h2 className="text-xl font-bold text-white">Administrator Access Required</h2>
        <p className="text-slate-400 text-sm max-w-md">
          This portal is restricted to platform administrators. Redirecting you to the home page...
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto py-12 px-4 sm:px-6 lg:px-8 space-y-8">
      {/* Title */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-6 border-b border-card-border">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-blurple/15 border border-blurple/30 text-blurple text-xs font-semibold uppercase tracking-wider mb-2">
            <ShieldCheck className="w-3.5 h-3.5" /> Platform Administration
          </div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">Admin & Inventory Command</h1>
          <p className="text-slate-400 text-sm mt-1">
            Manage pre-warmed turnkey token inventory, campaign promo discounts, gift voucher codes, and manual grants.
          </p>
        </div>
      </div>

      {message && (
        <div className="p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-medium flex items-center gap-2">
          <CheckCircle2 className="w-5 h-5 shrink-0" />
          <span>{message}</span>
        </div>
      )}

      {error && (
        <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-xs font-medium flex items-center gap-2">
          <AlertCircle className="w-5 h-5 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Navigation Tabs */}
      <div className="flex rounded-xl bg-slate-900 p-1 border border-slate-800 max-w-xl">
        <button
          type="button"
          onClick={() => setActiveTab('pool')}
          className={`flex-1 py-2 px-3 rounded-lg text-xs font-bold transition-all flex items-center justify-center gap-1.5 ${
            activeTab === 'pool' ? 'bg-blurple text-white shadow-md' : 'text-slate-400 hover:text-white'
          }`}
        >
          <Zap className="w-3.5 h-3.5" /> Token Pool
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('promo')}
          className={`flex-1 py-2 px-3 rounded-lg text-xs font-bold transition-all flex items-center justify-center gap-1.5 ${
            activeTab === 'promo' ? 'bg-blurple text-white shadow-md' : 'text-slate-400 hover:text-white'
          }`}
        >
          <Tag className="w-3.5 h-3.5" /> Promo Codes
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('voucher')}
          className={`flex-1 py-2 px-3 rounded-lg text-xs font-bold transition-all flex items-center justify-center gap-1.5 ${
            activeTab === 'voucher' ? 'bg-blurple text-white shadow-md' : 'text-slate-400 hover:text-white'
          }`}
        >
          <Gift className="w-3.5 h-3.5" /> Vouchers
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('grant')}
          className={`flex-1 py-2 px-3 rounded-lg text-xs font-bold transition-all flex items-center justify-center gap-1.5 ${
            activeTab === 'grant' ? 'bg-blurple text-white shadow-md' : 'text-slate-400 hover:text-white'
          }`}
        >
          <UserCheck className="w-3.5 h-3.5" /> Admin Grant
        </button>
      </div>

      {/* Tab 1: Pre-warmed Token Inventory Pool */}
      {activeTab === 'pool' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {['music', 'moderation', 'game'].map((category) => {
              const categoryStats = stats[category] || { available: 0, assigned: 0 };
              return (
                <div key={category} className="p-6 rounded-2xl bg-card border border-card-border space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
                      {category} Inventory
                    </span>
                    <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse" />
                  </div>
                  <div className="flex items-baseline gap-2">
                    <span className="text-3xl font-extrabold text-white">
                      {fetchingStats ? '...' : categoryStats.available ?? 0}
                    </span>
                    <span className="text-xs text-slate-400">available</span>
                  </div>
                  <div className="text-[11px] text-slate-500 pt-2 border-t border-slate-800 flex justify-between">
                    <span>Active Deployments:</span>
                    <span className="font-mono text-slate-300">{categoryStats.assigned ?? 0}</span>
                  </div>
                </div>
              );
            })}
          </div>

          {/* Add Token Form */}
          <div className="p-6 rounded-2xl bg-card border border-card-border space-y-4">
            <h3 className="text-sm font-bold text-white uppercase tracking-wider flex items-center gap-2">
              <Key className="w-4 h-4 text-amber-400" /> Ingest Pre-warmed Token into Vault
            </h3>
            <p className="text-xs text-slate-400">
              Tokens are immediately written to HashiCorp Vault under <code>secret/data/bots/pool/&#123;pool_id&#125;</code> and marked <code>available</code> for turnkey delivery.
            </p>

            <form onSubmit={handleAddToken} className="space-y-4 pt-2">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label className="text-xs text-slate-400 block mb-1">Bot Type</label>
                  <select
                    value={newBotType}
                    onChange={(e) => setNewBotType(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                  >
                    <option value="music">Music Bot</option>
                    <option value="moderation">Moderation Bot</option>
                    <option value="game">RPG Bot</option>
                  </select>
                </div>

                <div>
                  <label className="text-xs text-slate-400 block mb-1">Discord Application Client ID</label>
                  <input
                    type="text"
                    value={newClientId}
                    onChange={(e) => setNewClientId(e.target.value)}
                    placeholder="123456789012345678"
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-blurple"
                  />
                </div>

                <div>
                  <label className="text-xs text-slate-400 block mb-1">Discord Bot Token</label>
                  <input
                    type="password"
                    value={newToken}
                    onChange={(e) => setNewToken(e.target.value)}
                    placeholder="MTMxMj..."
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-blurple"
                  />
                </div>
              </div>

              <button
                type="submit"
                disabled={loading || !newClientId.trim() || !newToken.trim()}
                className="py-2.5 px-6 rounded-xl bg-blurple hover:bg-blurple-hover disabled:bg-slate-800 disabled:text-slate-500 text-white text-xs font-bold flex items-center gap-2 shadow-md shadow-blurple/20 transition-all"
              >
                {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Plus className="w-4 h-4" />}
                Add Token to Vault Pool
              </button>
            </form>
          </div>
        </div>
      )}

      {/* Tab 2: Promo Codes */}
      {activeTab === 'promo' && (
        <div className="p-6 rounded-2xl bg-card border border-card-border space-y-6">
          <div className="space-y-1">
            <h3 className="text-sm font-bold text-white uppercase tracking-wider flex items-center gap-2">
              <Tag className="w-4 h-4 text-blurple" /> Create Marketing Promo Code
            </h3>
            <p className="text-xs text-slate-400">
              Create discount coupons validated by billing-svc during checkout.
            </p>
          </div>

          <form onSubmit={handleCreatePromo} className="space-y-4 max-w-xl">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-xs text-slate-400 block mb-1">Code</label>
                <input
                  type="text"
                  value={promoCode}
                  onChange={(e) => setPromoCode(e.target.value)}
                  placeholder="e.g. FLASH50"
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white uppercase font-mono focus:outline-none focus:border-blurple"
                />
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">Discount Type</label>
                <select
                  value={promoDiscountType}
                  onChange={(e) => setPromoDiscountType(e.target.value as any)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                >
                  <option value="percentage">Percentage (%)</option>
                  <option value="fixed">Fixed Cents ($)</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">
                  Discount Value ({promoDiscountType === 'percentage' ? '%' : 'Cents'})
                </label>
                <input
                  type="number"
                  value={promoDiscountValue}
                  onChange={(e) => setPromoDiscountValue(Number(e.target.value))}
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                />
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">Max Redemptions</label>
                <input
                  type="number"
                  value={promoMaxUses}
                  onChange={(e) => setPromoMaxUses(Number(e.target.value))}
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading || !promoCode.trim()}
              className="py-2.5 px-6 rounded-xl bg-blurple hover:bg-blurple-hover disabled:bg-slate-800 disabled:text-slate-500 text-white text-xs font-bold transition-all shadow-md shadow-blurple/20 flex items-center gap-2"
            >
              {loading && <Loader2 className="w-4 h-4 animate-spin" />}
              Publish Promo Code
            </button>
          </form>
        </div>
      )}

      {/* Tab 3: Gift Vouchers */}
      {activeTab === 'voucher' && (
        <div className="p-6 rounded-2xl bg-card border border-card-border space-y-6">
          <div className="space-y-1">
            <h3 className="text-sm font-bold text-white uppercase tracking-wider flex items-center gap-2">
              <Gift className="w-4 h-4 text-emerald-400" /> Generate Gift Voucher
            </h3>
            <p className="text-xs text-slate-400">
              Creates single-use codes that can be redeemed without payment details for 100% free activations.
            </p>
          </div>

          <form onSubmit={handleCreateVoucher} className="space-y-4 max-w-xl">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-xs text-slate-400 block mb-1">Custom Code (Optional)</label>
                <input
                  type="text"
                  value={voucherCode}
                  onChange={(e) => setVoucherCode(e.target.value)}
                  placeholder="Leave empty for auto-generated"
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-blurple"
                />
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">Bot Type</label>
                <select
                  value={voucherBotType}
                  onChange={(e) => setVoucherBotType(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                >
                  <option value="music">Music</option>
                  <option value="moderation">Moderation</option>
                  <option value="game">Game</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">Plan Identifier</label>
                <input
                  type="text"
                  value={voucherPlanId}
                  onChange={(e) => setVoucherPlanId(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                />
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">Duration (Days)</label>
                <input
                  type="number"
                  value={voucherDays}
                  onChange={(e) => setVoucherDays(Number(e.target.value))}
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="py-2.5 px-6 rounded-xl bg-blurple hover:bg-blurple-hover disabled:bg-slate-800 disabled:text-slate-500 text-white text-xs font-bold transition-all shadow-md shadow-blurple/20 flex items-center gap-2"
            >
              {loading && <Loader2 className="w-4 h-4 animate-spin" />}
              Generate Voucher Code
            </button>
          </form>
        </div>
      )}

      {/* Tab 4: Direct Admin Grant */}
      {activeTab === 'grant' && (
        <div className="p-6 rounded-2xl bg-card border border-card-border space-y-6">
          <div className="space-y-1">
            <h3 className="text-sm font-bold text-white uppercase tracking-wider flex items-center gap-2">
              <UserCheck className="w-4 h-4 text-indigo-400" /> Direct Admin Grant
            </h3>
            <p className="text-xs text-slate-400">
              Immediately create and activate a subscription for any Discord User ID & Guild ID without requiring payment or checkout.
            </p>
          </div>

          <form onSubmit={handleAdminGrant} className="space-y-4 max-w-xl">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-xs text-slate-400 block mb-1">Discord User ID</label>
                <input
                  type="text"
                  value={grantUserId}
                  onChange={(e) => setGrantUserId(e.target.value)}
                  placeholder="123456789012345678"
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-blurple"
                />
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">Discord Guild ID</label>
                <input
                  type="text"
                  value={grantGuildId}
                  onChange={(e) => setGrantGuildId(e.target.value)}
                  placeholder="987654321098765432"
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-blurple"
                />
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">Bot Type</label>
                <select
                  value={grantBotType}
                  onChange={(e) => setGrantBotType(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                >
                  <option value="music">Music</option>
                  <option value="moderation">Moderation</option>
                  <option value="game">Game</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">Plan ID</label>
                <input
                  type="text"
                  value={grantPlanId}
                  onChange={(e) => setGrantPlanId(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                />
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">Duration (Days)</label>
                <input
                  type="number"
                  value={grantDurationDays}
                  onChange={(e) => setGrantDurationDays(Number(e.target.value))}
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading || !grantUserId.trim() || !grantGuildId.trim()}
              className="py-2.5 px-6 rounded-xl bg-blurple hover:bg-blurple-hover disabled:bg-slate-800 disabled:text-slate-500 text-white text-xs font-bold transition-all shadow-md shadow-blurple/20 flex items-center gap-2"
            >
              {loading && <Loader2 className="w-4 h-4 animate-spin" />}
              Grant Subscription Instantly
            </button>
          </form>
        </div>
      )}
    </div>
  );
}
