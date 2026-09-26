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
  Cpu,
  Loader2,
  LogIn,
  AlertCircle,
} from 'lucide-react';
import { getCatalogBots, checkTokenPoolAvailable } from '@/lib/api';
import { BotTemplate, SubscriptionPlan } from '@/lib/types';
import { useAuth } from '@/contexts/AuthContext';

export default function StorePage() {
  return (
    <Suspense
      fallback={
        <div className="max-w-7xl mx-auto py-12 px-4 text-center text-slate-400 text-sm flex items-center justify-center gap-2">
          <Loader2 className="w-5 h-5 animate-spin text-blurple" /> Loading Store...
        </div>
      }
    >
      <StoreContent />
    </Suspense>
  );
}

function StoreContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { user, guilds, selectedGuild, selectGuild, loginWithDiscord } = useAuth();

  const [bots, setBots] = useState<BotTemplate[]>([]);
  const [selectedBot, setSelectedBot] = useState<BotTemplate | null>(null);
  const [selectedPlan, setSelectedPlan] = useState<SubscriptionPlan | null>(null);
  const [targetGuildId, setTargetGuildId] = useState<string>('');
  const [customInstanceLabel, setCustomInstanceLabel] = useState<string>('');
  const [zeroSetup, setZeroSetup] = useState<boolean>(true);
  const [poolAvailable, setPoolAvailable] = useState<boolean>(true);
  const [poolCount, setPoolCount] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch bot catalog from catalog-svc
  useEffect(() => {
    getCatalogBots()
      .then((data) => {
        setBots(data);
        if (data.length > 0) {
          const requestedBotId = searchParams.get('bot');
          const matched = data.find((b) => b.id === requestedBotId || b.slug === requestedBotId);
          const activeBot = matched || data[0];
          setSelectedBot(activeBot);

          const requestedPlanId = searchParams.get('plan');
          const matchedPlan = activeBot.plans?.find((p) => p.id === requestedPlanId);
          setSelectedPlan(matchedPlan || activeBot.plans?.[activeBot.plans.length - 1] || activeBot.plans?.[0] || null);
        }
      })
      .catch((err) => {
        console.error('Failed fetching catalog:', err);
        setError('Failed to fetch bot templates from catalog service.');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [searchParams]);

  // Sync target guild with auth context
  useEffect(() => {
    if (selectedGuild) {
      setTargetGuildId(selectedGuild.id);
    } else if (guilds.length > 0) {
      setTargetGuildId(guilds[0].id);
    }
  }, [selectedGuild, guilds]);

  // Check token pool availability when bot selection changes
  useEffect(() => {
    if (selectedBot) {
      if (selectedBot.plans && selectedBot.plans.length > 0 && !selectedPlan) {
        setSelectedPlan(selectedBot.plans[selectedBot.plans.length - 1]);
      }

      checkTokenPoolAvailable(selectedBot.category).then((res) => {
        setPoolAvailable(res.is_available);
        setPoolCount(res.available_count);
      });
    }
  }, [selectedBot, selectedPlan]);

  const basePriceCents = selectedPlan?.price_cents || 0;
  const zeroSetupFeeCents = zeroSetup ? 299 : 0;
  const totalPriceCents = basePriceCents + zeroSetupFeeCents;

  const handleProceedToCheckout = () => {
    if (!selectedBot || !selectedPlan) return;
    if (!targetGuildId) {
      setError('Please select or specify a target Discord server before proceeding.');
      return;
    }

    const params = new URLSearchParams({
      bot: selectedBot.id,
      botType: selectedBot.category,
      plan: selectedPlan.id,
      guild: targetGuildId,
      instanceLabel: customInstanceLabel.trim() || 'Default',
      zeroSetup: zeroSetup ? 'true' : 'false',
    });
    router.push(`/checkout?${params.toString()}`);
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto py-24 px-4 text-center space-y-4">
        <Loader2 className="w-8 h-8 animate-spin text-blurple mx-auto" />
        <p className="text-slate-400 text-sm">Loading bot catalog from Catalog Service...</p>
      </div>
    );
  }

  if (error || bots.length === 0) {
    return (
      <div className="max-w-xl mx-auto py-24 px-4 text-center space-y-4">
        <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-300 text-sm flex items-center justify-center gap-2">
          <AlertCircle className="w-5 h-5 shrink-0" />
          <span>{error || 'No bot templates found in catalog.'}</span>
        </div>
      </div>
    );
  }

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
            {bots.map((bot) => {
              const isSelected = selectedBot?.id === bot.id;
              const isMusic = bot.category === 'music';
              const isMod = bot.category === 'moderation';
              const isGame = bot.category === 'game';
              const Icon = isMusic ? Music : isMod ? Shield : isGame ? Gamepad2 : Cpu;
              const colorClass = isMusic ? 'text-cyan-400' : isMod ? 'text-amber-400' : 'text-purple-400';

              return (
                <button
                  key={bot.id}
                  type="button"
                  onClick={() => {
                    setSelectedBot(bot);
                    if (bot.plans && bot.plans.length > 0) {
                      setSelectedPlan(bot.plans[bot.plans.length - 1]);
                    }
                  }}
                  className={`w-full p-4 rounded-2xl text-left border transition-all flex items-start gap-4 ${
                    isSelected
                      ? 'bg-blurple/15 border-blurple shadow-md'
                      : 'bg-card border-card-border hover:border-slate-700'
                  }`}
                >
                  <div
                    className={`p-3 rounded-xl shrink-0 ${
                      isSelected ? 'bg-blurple text-white' : 'bg-slate-800 ' + colorClass
                    }`}
                  >
                    <Icon className="w-5 h-5" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <h4 className="text-sm font-bold text-white truncate">{bot.name}</h4>
                      <span className="text-[10px] font-mono uppercase px-2 py-0.5 rounded bg-slate-800 text-slate-300">
                        {bot.category}
                      </span>
                    </div>
                    <p className="text-xs text-slate-400 mt-1 line-clamp-2">{bot.description}</p>
                  </div>
                </button>
              );
            })}
          </div>

          {/* Target Discord Server Selection */}
          <div className="p-5 rounded-2xl bg-card border border-card-border space-y-3">
            <h4 className="text-xs font-bold uppercase tracking-wider text-slate-300 flex items-center gap-1.5">
              <Server className="w-3.5 h-3.5 text-blurple" /> Target Discord Server
            </h4>

            {user ? (
              guilds.length > 0 ? (
                <div className="space-y-2">
                  <select
                    value={targetGuildId}
                    onChange={(e) => {
                      setTargetGuildId(e.target.value);
                      const g = guilds.find((x) => x.id === e.target.value);
                      if (g) selectGuild(g);
                    }}
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2.5 text-xs text-white focus:outline-none focus:border-blurple"
                  >
                    {guilds.map((g) => (
                      <option key={g.id} value={g.id}>
                        {g.name} {g.owner ? '(Owner)' : '(Admin)'}
                      </option>
                    ))}
                  </select>
                  <p className="text-[11px] text-slate-400">
                    Fetched from your active Discord account. Only servers with Manage Server permissions are listed.
                  </p>
                </div>
              ) : (
                <div className="space-y-2">
                  <input
                    type="text"
                    value={targetGuildId}
                    onChange={(e) => setTargetGuildId(e.target.value)}
                    placeholder="Enter Guild ID (e.g. 1122334455)"
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white font-mono focus:outline-none focus:border-blurple"
                  />
                  <p className="text-[11px] text-slate-400">
                    No servers detected from OAuth. Enter your target Discord Server ID directly.
                  </p>
                </div>
              )
            ) : (
              <div className="p-3 rounded-xl bg-slate-900/80 border border-slate-800 space-y-2.5">
                <p className="text-xs text-slate-300">Log in to automatically select from your Discord servers:</p>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => loginWithDiscord()}
                    className="w-full py-1.5 px-3 rounded-lg bg-blurple hover:bg-blurple-hover text-white text-xs font-semibold flex items-center justify-center gap-1.5"
                  >
                    <LogIn className="w-3.5 h-3.5" /> Sign in with Discord
                  </button>
                </div>
              </div>
            )}

            {/* Instance Label for Multi-Subscription Tenancy */}
            <div className="pt-2 border-t border-slate-800">
              <label className="text-[11px] text-slate-400 block mb-1 font-medium">
                Instance Label (Multi-Bot Tag)
              </label>
              <input
                type="text"
                value={customInstanceLabel}
                onChange={(e) => setCustomInstanceLabel(e.target.value)}
                placeholder="e.g. Main Lobby, VIP DJ, Defense Shield"
                className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
              />
              <span className="text-[10px] text-slate-500 mt-1 block">
                Allows running multiple bots of the same type without collisions.
              </span>
            </div>
          </div>
        </div>

        {/* Middle & Right Columns: Plan Configurator & Delivery Method */}
        <div className="lg:col-span-2 space-y-6">
          <div className="space-y-4">
            <h3 className="text-sm font-bold uppercase tracking-wider text-slate-400">2. Select Subscription Tier</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {selectedBot?.plans?.map((plan) => {
                const isSelected = selectedPlan?.id === plan.id;
                const isDedicated = plan.is_dedicated;

                return (
                  <button
                    key={plan.id}
                    type="button"
                    onClick={() => setSelectedPlan(plan)}
                    className={`p-6 rounded-2xl text-left border flex flex-col justify-between transition-all ${
                      isSelected
                        ? 'bg-slate-900 border-blurple ring-1 ring-blurple shadow-xl'
                        : 'bg-card border-card-border hover:border-slate-700'
                    }`}
                  >
                    <div className="space-y-4">
                      <div className="flex items-center justify-between">
                        <span
                          className={`text-xs font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${
                            isDedicated
                              ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                              : 'bg-slate-800 text-slate-400'
                          }`}
                        >
                          {isDedicated ? 'Dedicated Pod' : 'Shared Host'}
                        </span>
                        {isSelected && <CheckCircle2 className="w-5 h-5 text-blurple" />}
                      </div>

                      <div>
                        <h4 className="text-base font-bold text-white">{plan.name}</h4>
                        <p className="text-xs text-slate-400 mt-1">{plan.description}</p>
                      </div>

                      <div className="pt-2">
                        <span className="text-2xl font-black text-white">
                          {plan.price_cents === 0 ? 'Free' : `$${(plan.price_cents / 100).toFixed(2)}`}
                        </span>
                        <span className="text-xs text-slate-400 font-medium ml-1">/ month</span>
                      </div>

                      <div className="space-y-2 pt-2 border-t border-slate-800">
                        {plan.features?.map((f, i) => (
                          <div key={i} className="text-xs text-slate-300 flex items-center gap-2">
                            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                            <span>{f}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Turnkey Zero-Setup Toggle */}
          <div className="p-6 rounded-2xl bg-card border border-card-border space-y-4">
            <div className="flex items-start justify-between gap-4">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <h4 className="text-sm font-bold text-white flex items-center gap-2">
                    <Zap className="w-4 h-4 text-amber-400" />
                    Zero-Setup Turnkey Delivery
                  </h4>
                  <span className="text-[10px] font-mono uppercase px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/20">
                    Recommended
                  </span>
                </div>
                <p className="text-xs text-slate-400 leading-relaxed max-w-xl">
                  Don't want to create Discord applications, copy bot tokens, or configure intents? Our admin token pool automatically provides pre-warmed credentials. 1-click deployment straight into your server.
                </p>
                <div className="text-[11px] text-slate-400 flex items-center gap-2 pt-1">
                  <span className={`w-2 h-2 rounded-full ${poolAvailable ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'}`} />
                  <span>
                    Pre-warmed Inventory: <strong>{poolCount} available tokens</strong> for {selectedBot?.name}
                  </span>
                </div>
              </div>

              <label className="relative inline-flex items-center cursor-pointer shrink-0 mt-1">
                <input
                  type="checkbox"
                  checked={zeroSetup}
                  onChange={(e) => setZeroSetup(e.target.checked)}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-slate-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blurple"></div>
              </label>
            </div>
          </div>

          {/* Order Summary Card */}
          <div className="p-6 rounded-2xl bg-slate-900/90 border border-slate-800 space-y-4">
            <div className="flex items-center justify-between text-sm">
              <span className="text-slate-400">Monthly Subscription ({selectedPlan?.name}):</span>
              <span className="font-mono text-white font-bold">
                {basePriceCents === 0 ? '$0.00' : `$${(basePriceCents / 100).toFixed(2)}`}
              </span>
            </div>

            {zeroSetup && (
              <div className="flex items-center justify-between text-sm">
                <span className="text-slate-400 flex items-center gap-1.5">
                  <Zap className="w-3.5 h-3.5 text-amber-400" /> Pre-warmed Token Allocation:
                </span>
                <span className="font-mono text-white font-bold">$2.99</span>
              </div>
            )}

            <div className="pt-3 border-t border-slate-800 flex items-center justify-between">
              <div>
                <span className="text-xs text-slate-400 block font-medium">Total Billed Today</span>
                <span className="text-2xl font-black text-white">
                  {totalPriceCents === 0 ? 'Free' : `$${(totalPriceCents / 100).toFixed(2)}`}
                </span>
              </div>

              <button
                type="button"
                onClick={handleProceedToCheckout}
                className="py-3 px-6 rounded-xl bg-blurple hover:bg-blurple-hover text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-blurple/25 transition-all"
              >
                Proceed to Checkout <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
