'use client';

import React, { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import {
  ShoppingBag,
  Zap,
  CheckCircle2,
  Server,
  ArrowRight,
  Shield,
  Music,
  Gamepad2,
  Sparkles,
  HelpCircle,
} from 'lucide-react';
import { MOCK_BOTS, MOCK_GUILDS, checkTokenPoolAvailable } from '@/lib/api';
import { BotTemplate, SubscriptionPlan } from '@/lib/types';

export default function StorePage() {
  return (
    <Suspense fallback={<div className="max-w-7xl mx-auto py-12 px-4 text-center text-slate-400 text-sm">Loading Store...</div>}>
      <StoreContent />
    </Suspense>
  );
}

function StoreContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialBotId = searchParams.get('bot') || MOCK_BOTS[0].id;

  const [selectedBot, setSelectedBot] = useState<BotTemplate>(
    MOCK_BOTS.find((b) => b.id === initialBotId) || MOCK_BOTS[0]
  );
  const [selectedPlan, setSelectedPlan] = useState<SubscriptionPlan>(
    selectedBot.plans?.[1] || selectedBot.plans?.[0]!
  );
  const [selectedGuild, setSelectedGuild] = useState<string>(MOCK_GUILDS[0].id);
  const [zeroSetup, setZeroSetup] = useState<boolean>(true);
  const [poolAvailable, setPoolAvailable] = useState<boolean>(true);
  const [poolCount, setPoolCount] = useState<number>(5);

  useEffect(() => {
    // When bot changes, reset plan selection
    if (selectedBot.plans && selectedBot.plans.length > 0) {
      setSelectedPlan(selectedBot.plans[selectedBot.plans.length - 1]);
    }

    // Check pre-warmed token pool availability for selected bot
    checkTokenPoolAvailable(selectedBot.category).then((res) => {
      setPoolAvailable(res.is_available);
      setPoolCount(res.available_count);
    });
  }, [selectedBot]);

  const basePriceCents = selectedPlan.price_cents;
  const zeroSetupFeeCents = zeroSetup ? 299 : 0;
  const totalPriceCents = basePriceCents + zeroSetupFeeCents;

  const handleProceedToCheckout = () => {
    const params = new URLSearchParams({
      bot: selectedBot.id,
      botType: selectedBot.category,
      plan: selectedPlan.id,
      guild: selectedGuild,
      zeroSetup: zeroSetup ? 'true' : 'false',
    });
    router.push(`/checkout?${params.toString()}`);
  };

  return (
    <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8 space-y-12">
      {/* Header */}
      <div className="space-y-3">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blurple/15 border border-blurple/30 text-blurple text-xs font-semibold uppercase tracking-wider">
          <ShoppingBag className="w-3.5 h-3.5" /> Bot Marketplace
        </div>
        <h1 className="text-3xl sm:text-4xl font-extrabold text-white tracking-tight">
          Select Your Discord Bot & Deployment Plan
        </h1>
        <p className="text-slate-400 text-sm max-w-2xl">
          Choose a bot template, pick your server, and customize deployment options. Turnkey Zero-Setup provides pre-configured bot credentials with zero developer hassle.
        </p>
      </div>

      {/* Main Grid: Bot Selector + Plan Configurator */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Column: Bot Template Selection */}
        <div className="space-y-4">
          <h3 className="text-sm font-bold uppercase tracking-wider text-slate-400">1. Select Bot Template</h3>
          <div className="space-y-3">
            {MOCK_BOTS.map((bot) => {
              const isSelected = selectedBot.id === bot.id;
              const isMusic = bot.category === 'music';
              const isMod = bot.category === 'moderation';
              return (
                <button
                  key={bot.id}
                  type="button"
                  onClick={() => setSelectedBot(bot)}
                  className={`w-full text-left p-4 rounded-xl border transition-all flex items-start gap-4 ${
                    isSelected
                      ? 'bg-blurple/10 border-blurple shadow-md'
                      : 'bg-card border-card-border hover:border-slate-700'
                  }`}
                >
                  <div
                    className={`w-10 h-10 rounded-lg flex items-center justify-center shrink-0 text-white ${
                      isMusic ? 'bg-purple-600' : isMod ? 'bg-blue-600' : 'bg-emerald-600'
                    }`}
                  >
                    {isMusic ? (
                      <Music className="w-5 h-5" />
                    ) : isMod ? (
                      <Shield className="w-5 h-5" />
                    ) : (
                      <Gamepad2 className="w-5 h-5" />
                    )}
                  </div>
                  <div>
                    <h4 className="text-sm font-bold text-white flex items-center gap-2">
                      {bot.name}
                      {isSelected && (
                        <span className="w-2 h-2 rounded-full bg-blurple" />
                      )}
                    </h4>
                    <p className="text-xs text-slate-400 line-clamp-2 mt-1">{bot.description}</p>
                  </div>
                </button>
              );
            })}
          </div>

          {/* Server Selector */}
          <div className="pt-6 border-t border-card-border space-y-3">
            <h3 className="text-sm font-bold uppercase tracking-wider text-slate-400">2. Target Discord Server</h3>
            <div className="space-y-2">
              <label className="text-xs text-slate-400">Select where the bot will be installed:</label>
              <select
                value={selectedGuild}
                onChange={(e) => setSelectedGuild(e.target.value)}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-white focus:outline-none focus:border-blurple"
              >
                {MOCK_GUILDS.map((g) => (
                  <option key={g.id} value={g.id}>
                    {g.name} (Admin)
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Center Column: Plan Tiers */}
        <div className="space-y-4">
          <h3 className="text-sm font-bold uppercase tracking-wider text-slate-400">3. Choose Subscription Tier</h3>
          <div className="space-y-4">
            {selectedBot.plans?.map((plan) => {
              const isSelected = selectedPlan.id === plan.id;
              return (
                <div
                  key={plan.id}
                  onClick={() => setSelectedPlan(plan)}
                  className={`cursor-pointer rounded-2xl p-5 border transition-all relative ${
                    isSelected
                      ? 'bg-slate-900/90 border-blurple shadow-lg ring-1 ring-blurple/50'
                      : 'bg-card border-card-border hover:border-slate-700'
                  }`}
                >
                  {plan.is_dedicated && (
                    <span className="absolute top-4 right-4 text-[10px] font-mono px-2 py-0.5 rounded-full bg-blurple/20 text-blurple border border-blurple/30 font-semibold uppercase">
                      Dedicated K8s Pod
                    </span>
                  )}

                  <div className="space-y-2">
                    <h4 className="text-base font-bold text-white">{plan.name}</h4>
                    <p className="text-xs text-slate-400">{plan.description}</p>
                    <div className="text-2xl font-black text-white pt-1">
                      ${plan.price_cents / 100}
                      <span className="text-xs font-normal text-slate-400"> / month</span>
                    </div>
                  </div>

                  <div className="pt-4 mt-4 border-t border-slate-800 space-y-1.5">
                    {plan.features.map((feature, idx) => (
                      <div key={idx} className="flex items-center gap-2 text-xs text-slate-300">
                        <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                        <span>{feature}</span>
                      </div>
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Right Column: Checkout Summary & Zero-Setup Addon */}
        <div className="space-y-4">
          <h3 className="text-sm font-bold uppercase tracking-wider text-slate-400">4. Turnkey Options & Order</h3>

          <div className="glass-panel rounded-2xl p-6 space-y-6 border border-card-border">
            {/* Zero-Setup Managed Token Pool Card */}
            <div
              onClick={() => poolAvailable && setZeroSetup(!zeroSetup)}
              className={`p-4 rounded-xl border transition-all cursor-pointer ${
                zeroSetup
                  ? 'bg-amber-500/10 border-amber-500/40'
                  : 'bg-slate-900/40 border-slate-800 hover:border-slate-700'
              }`}
            >
              <div className="flex items-start justify-between gap-3">
                <div className="flex items-center gap-2.5">
                  <div className="p-1.5 rounded-md bg-amber-500/20 text-amber-400">
                    <Zap className="w-4 h-4" />
                  </div>
                  <div>
                    <h5 className="text-xs font-bold text-white flex items-center gap-1.5">
                      Turnkey 0-Setup Deployment
                      <span className="text-[10px] px-1.5 py-0.2 rounded bg-amber-400/20 text-amber-300 font-mono">
                        +$2.99/mo
                      </span>
                    </h5>
                    <p className="text-[11px] text-slate-400 mt-1">
                      No Discord Developer Portal needed. Instant bot invite link generated right away.
                    </p>
                  </div>
                </div>
                <input
                  type="checkbox"
                  checked={zeroSetup}
                  onChange={(e) => setZeroSetup(e.target.checked)}
                  disabled={!poolAvailable}
                  className="mt-1 h-4 w-4 rounded border-slate-700 text-blurple focus:ring-blurple shrink-0"
                />
              </div>

              {poolAvailable ? (
                <div className="mt-3 flex items-center gap-1.5 text-[11px] text-emerald-400 font-medium">
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  <span>In Stock: {poolCount} pre-warmed tokens ready for immediate deployment</span>
                </div>
              ) : (
                <div className="mt-3 text-[11px] text-red-400 font-medium">
                  Sold out for this bot type. Self-setup token required.
                </div>
              )}
            </div>

            {/* Price Calculation Summary */}
            <div className="space-y-3 pt-4 border-t border-slate-800 text-xs">
              <div className="flex justify-between text-slate-400">
                <span>{selectedBot.name} ({selectedPlan.name})</span>
                <span className="text-white font-mono">${(basePriceCents / 100).toFixed(2)}</span>
              </div>
              {zeroSetup && (
                <div className="flex justify-between text-slate-400">
                  <span className="flex items-center gap-1">
                    <Zap className="w-3 h-3 text-amber-400" /> Turnkey 0-Setup Service
                  </span>
                  <span className="text-white font-mono">${(zeroSetupFeeCents / 100).toFixed(2)}</span>
                </div>
              )}
              <div className="flex justify-between text-base font-bold text-white pt-3 border-t border-slate-800">
                <span>Total Due Monthly</span>
                <span className="text-blurple font-mono text-xl">${(totalPriceCents / 100).toFixed(2)}</span>
              </div>
            </div>

            {/* Action Button */}
            <button
              type="button"
              onClick={handleProceedToCheckout}
              className="w-full py-3.5 px-4 rounded-xl bg-blurple hover:bg-blurple-hover text-white font-semibold text-sm flex items-center justify-center gap-2 shadow-lg shadow-blurple/25 transition-all"
            >
              Continue to Checkout <ArrowRight className="w-4 h-4" />
            </button>

            <div className="flex items-center justify-center gap-2 text-[11px] text-slate-500">
              <Shield className="w-3.5 h-3.5" /> Cancel anytime with 1-click in Dashboard or Discord
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
