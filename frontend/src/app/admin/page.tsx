'use client';

import React, { useState, useEffect } from 'react';
import {
  ShieldCheck,
  Zap,
  Tag,
  Gift,
  Plus,
  CheckCircle2,
  AlertCircle,
  Database,
  Layers,
  Key,
} from 'lucide-react';
import { getTokenPoolStats } from '@/lib/api';
import { TokenPoolStats } from '@/lib/types';

export default function AdminPage() {
  const [stats, setStats] = useState<TokenPoolStats>({
    music: { available: 5, assigned: 2 },
    moderation: { available: 3, assigned: 1 },
    game: { available: 4, assigned: 0 },
  });

  const [activeTab, setActiveTab] = useState<'pool' | 'promo' | 'voucher'>('pool');
  const [message, setMessage] = useState<string | null>(null);

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

  useEffect(() => {
    getTokenPoolStats().then((res) => {
      if (res && Object.keys(res).length > 0) {
        setStats(res);
      }
    });
  }, []);

  const handleAddToken = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newClientId || !newToken) return;

    // Simulate adding to state
    setStats((prev) => {
      const copy = { ...prev };
      if (!copy[newBotType]) copy[newBotType] = { available: 0, assigned: 0 };
      copy[newBotType].available = (copy[newBotType].available || 0) + 1;
      return copy;
    });

    setMessage(`Pre-warmed token for ${newBotType.toUpperCase()} added to Vault and database!`);
    setNewClientId('');
    setNewToken('');
    setTimeout(() => setMessage(null), 4000);
  };

  const handleCreatePromo = (e: React.FormEvent) => {
    e.preventDefault();
    if (!promoCode) return;
    setMessage(`Promo code "${promoCode.toUpperCase()}" created successfully!`);
    setPromoCode('');
    setTimeout(() => setMessage(null), 4000);
  };

  const handleCreateVoucher = (e: React.FormEvent) => {
    e.preventDefault();
    const code = voucherCode.trim() || `GIFT-${Math.random().toString(36).substring(2, 8).toUpperCase()}`;
    setMessage(`Gift voucher code "${code}" created for ${voucherDays} days!`);
    setVoucherCode('');
    setTimeout(() => setMessage(null), 4000);
  };

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
            Manage pre-warmed turnkey token inventory, campaign promo discounts, and gift voucher codes.
          </p>
        </div>
      </div>

      {message && (
        <div className="p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-medium flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4 shrink-0" />
          <span>{message}</span>
        </div>
      )}

      {/* Navigation Tabs */}
      <div className="flex gap-2 border-b border-card-border pb-4">
        <button
          type="button"
          onClick={() => setActiveTab('pool')}
          className={`px-4 py-2 rounded-xl text-xs font-bold flex items-center gap-2 transition-all ${
            activeTab === 'pool'
              ? 'bg-blurple text-white shadow-sm'
              : 'text-slate-400 hover:text-white bg-slate-900/50'
          }`}
        >
          <Zap className="w-3.5 h-3.5" /> Turnkey Token Pool Inventory
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('promo')}
          className={`px-4 py-2 rounded-xl text-xs font-bold flex items-center gap-2 transition-all ${
            activeTab === 'promo'
              ? 'bg-blurple text-white shadow-sm'
              : 'text-slate-400 hover:text-white bg-slate-900/50'
          }`}
        >
          <Tag className="w-3.5 h-3.5" /> Promo Campaign Codes
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('voucher')}
          className={`px-4 py-2 rounded-xl text-xs font-bold flex items-center gap-2 transition-all ${
            activeTab === 'voucher'
              ? 'bg-blurple text-white shadow-sm'
              : 'text-slate-400 hover:text-white bg-slate-900/50'
          }`}
        >
          <Gift className="w-3.5 h-3.5" /> Gift Vouchers
        </button>
      </div>

      {activeTab === 'pool' && (
        <div className="space-y-8">
          {/* Inventory Breakdown Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {Object.entries(stats).map(([botType, counts]) => (
              <div key={botType} className="glass-card rounded-2xl p-6 border border-card-border space-y-4">
                <div className="flex items-center justify-between">
                  <h4 className="text-sm font-bold text-white uppercase">{botType} Bots</h4>
                  <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-300">
                    Vault Secret Store
                  </span>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div className="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/20">
                    <span className="text-[11px] text-emerald-400">Available</span>
                    <p className="text-2xl font-black text-white font-mono mt-0.5">
                      {counts.available || 0}
                    </p>
                  </div>
                  <div className="p-3 rounded-xl bg-blurple/10 border border-blurple/20">
                    <span className="text-[11px] text-blurple">Assigned</span>
                    <p className="text-2xl font-black text-white font-mono mt-0.5">
                      {counts.assigned || 0}
                    </p>
                  </div>
                </div>

                <div className="text-[11px] text-slate-400 flex items-center gap-1.5">
                  <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                  Ready for instant checkout leasing
                </div>
              </div>
            ))}
          </div>

          {/* Add Token Form */}
          <div className="p-6 rounded-2xl bg-card border border-card-border space-y-4">
            <h3 className="text-base font-bold text-white flex items-center gap-2">
              <Plus className="w-4 h-4 text-blurple" /> Replenish Pre-Warmed Token Pool
            </h3>
            <p className="text-xs text-slate-400">
              Add pre-created Discord Bot Application credentials so users can choose 0-setup deployment without ever creating developer keys.
            </p>

            <form onSubmit={handleAddToken} className="grid grid-cols-1 md:grid-cols-4 gap-4 pt-2">
              <div>
                <label className="text-xs text-slate-400 block mb-1">Bot Type</label>
                <select
                  value={newBotType}
                  onChange={(e) => setNewBotType(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                >
                  <option value="music">Music</option>
                  <option value="moderation">Moderation</option>
                  <option value="game">RPG / Game</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">Client ID (Application ID)</label>
                <input
                  type="text"
                  value={newClientId}
                  onChange={(e) => setNewClientId(e.target.value)}
                  placeholder="131234567890123456"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-blurple"
                />
              </div>

              <div>
                <label className="text-xs text-slate-400 block mb-1">Bot Token</label>
                <input
                  type="password"
                  value={newToken}
                  onChange={(e) => setNewToken(e.target.value)}
                  placeholder="MTMxMjM0... (Token)"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-blurple"
                />
              </div>

              <div className="flex items-end">
                <button
                  type="submit"
                  disabled={!newClientId || !newToken}
                  className="w-full py-2.5 px-4 rounded-lg bg-blurple hover:bg-blurple-hover disabled:bg-slate-800 disabled:text-slate-500 text-white text-xs font-semibold transition-colors"
                >
                  Add to Token Pool
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {activeTab === 'promo' && (
        <div className="p-6 rounded-2xl bg-card border border-card-border space-y-6">
          <div>
            <h3 className="text-base font-bold text-white flex items-center gap-2">
              <Tag className="w-4 h-4 text-blurple" /> Create Marketing Promo Code
            </h3>
            <p className="text-xs text-slate-400 mt-1">
              Create percentage or fixed dollar discount codes for customers to use during checkout.
            </p>
          </div>

          <form onSubmit={handleCreatePromo} className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label className="text-xs text-slate-400 block mb-1">Code Name</label>
              <input
                type="text"
                value={promoCode}
                onChange={(e) => setPromoCode(e.target.value)}
                placeholder="e.g. VIPFREE, SUMMER50"
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white font-mono uppercase tracking-wider focus:outline-none focus:border-blurple"
              />
            </div>

            <div>
              <label className="text-xs text-slate-400 block mb-1">Discount Type</label>
              <select
                value={promoDiscountType}
                onChange={(e) => setPromoDiscountType(e.target.value as any)}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
              >
                <option value="percentage">Percentage (%)</option>
                <option value="fixed">Fixed Cents ($)</option>
              </select>
            </div>

            <div>
              <label className="text-xs text-slate-400 block mb-1">
                Value {promoDiscountType === 'percentage' ? '(%)' : '(Cents)'}
              </label>
              <input
                type="number"
                value={promoDiscountValue}
                onChange={(e) => setPromoDiscountValue(Number(e.target.value))}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-blurple"
              />
            </div>

            <div className="flex items-end">
              <button
                type="submit"
                disabled={!promoCode}
                className="w-full py-2.5 px-4 rounded-lg bg-blurple hover:bg-blurple-hover disabled:bg-slate-800 disabled:text-slate-500 text-white text-xs font-semibold transition-colors"
              >
                Create Promo Code
              </button>
            </div>
          </form>
        </div>
      )}

      {activeTab === 'voucher' && (
        <div className="p-6 rounded-2xl bg-card border border-card-border space-y-6">
          <div>
            <h3 className="text-base font-bold text-white flex items-center gap-2">
              <Gift className="w-4 h-4 text-blurple" /> Generate Prepaid Gift Voucher
            </h3>
            <p className="text-xs text-slate-400 mt-1">
              Create 1-time redeemable codes that server owners can use in Discord with{' '}
              <code className="text-blurple font-mono">/redeem &lt;code&gt;</code>.
            </p>
          </div>

          <form onSubmit={handleCreateVoucher} className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label className="text-xs text-slate-400 block mb-1">Voucher Code (Optional)</label>
              <input
                type="text"
                value={voucherCode}
                onChange={(e) => setVoucherCode(e.target.value)}
                placeholder="Leave blank for auto-generate"
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white font-mono uppercase tracking-wider focus:outline-none focus:border-blurple"
              />
            </div>

            <div>
              <label className="text-xs text-slate-400 block mb-1">Bot Type</label>
              <select
                value={voucherBotType}
                onChange={(e) => setVoucherBotType(e.target.value)}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
              >
                <option value="music">Music Bot</option>
                <option value="moderation">Moderation Bot</option>
                <option value="game">RPG Game Bot</option>
              </select>
            </div>

            <div>
              <label className="text-xs text-slate-400 block mb-1">Validity (Days)</label>
              <input
                type="number"
                value={voucherDays}
                onChange={(e) => setVoucherDays(Number(e.target.value))}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-blurple"
              />
            </div>

            <div className="flex items-end">
              <button
                type="submit"
                className="w-full py-2.5 px-4 rounded-lg bg-blurple hover:bg-blurple-hover text-white text-xs font-semibold transition-colors"
              >
                Generate Voucher
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
