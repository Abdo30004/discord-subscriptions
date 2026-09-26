'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import {
  Sparkles,
  Zap,
  Shield,
  Music,
  Gamepad2,
  Cpu,
  ArrowRight,
  CheckCircle2,
  Terminal,
  Activity,
  Layers,
  Lock,
  Loader2,
} from 'lucide-react';
import { getCatalogBots } from '@/lib/api';
import { BotTemplate } from '@/lib/types';

export default function HomePage() {
  const [bots, setBots] = useState<BotTemplate[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getCatalogBots()
      .then((data) => {
        setBots(data);
        setError(null);
      })
      .catch((err) => {
        console.error('Failed to load catalog:', err);
        setError('Unable to load live catalog from catalog-svc. Ensure the service is running.');
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  return (
    <div className="space-y-24 py-12 px-4 sm:px-6 lg:px-8 max-w-7xl mx-auto">
      {/* Hero Section */}
      <section className="text-center space-y-6 pt-8 pb-4 relative">
        <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-blurple/15 border border-blurple/30 text-blurple text-xs font-semibold tracking-wide uppercase">
          <Sparkles className="w-3.5 h-3.5" /> Next-Generation Discord Cloud Infrastructure
        </div>

        <h1 className="text-4xl sm:text-6xl font-extrabold text-white tracking-tight leading-tight max-w-4xl mx-auto">
          Deploy Dedicated Discord Bots with{' '}
          <span className="bg-gradient-to-r from-blurple via-indigo-400 to-cyan-400 bg-clip-text text-transparent">
            Zero-Setup
          </span>
        </h1>

        <p className="text-lg sm:text-xl text-slate-400 max-w-2xl mx-auto leading-relaxed">
          The all-in-one subscription marketplace and Kubernetes orchestrator. Non-technical users get pre-warmed instant bots; power users get dedicated cluster pods with custom tokens and full persona customization.
        </p>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
          <Link
            href="/store"
            className="w-full sm:w-auto px-8 py-3.5 rounded-xl bg-blurple hover:bg-blurple-hover text-white font-semibold flex items-center justify-center gap-2 transition-all shadow-lg shadow-blurple/25"
          >
            Explore Bot Catalog <ArrowRight className="w-4 h-4" />
          </Link>
          <Link
            href="/dashboard"
            className="w-full sm:w-auto px-8 py-3.5 rounded-xl bg-slate-800/80 hover:bg-slate-700 text-slate-200 font-semibold flex items-center justify-center gap-2 border border-slate-700 transition-all"
          >
            Manage Existing Servers
          </Link>
        </div>

        {/* Feature Badges */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 pt-12 max-w-4xl mx-auto text-left">
          <div className="p-4 rounded-xl bg-card border border-card-border flex items-start gap-3">
            <div className="p-2 rounded-lg bg-amber-500/10 text-amber-400">
              <Zap className="w-5 h-5" />
            </div>
            <div>
              <h4 className="text-sm font-semibold text-white">0-Setup Turnkey</h4>
              <p className="text-xs text-slate-400 mt-0.5">Pre-warmed bot pool. No API keys or Dev Portal.</p>
            </div>
          </div>

          <div className="p-4 rounded-xl bg-card border border-card-border flex items-start gap-3">
            <div className="p-2 rounded-lg bg-emerald-500/10 text-emerald-400">
              <Cpu className="w-5 h-5" />
            </div>
            <div>
              <h4 className="text-sm font-semibold text-white">K8s Isolation</h4>
              <p className="text-xs text-slate-400 mt-0.5">Single-tenant pods, 320kbps audio, zero lag.</p>
            </div>
          </div>

          <div className="p-4 rounded-xl bg-card border border-card-border flex items-start gap-3">
            <div className="p-2 rounded-lg bg-blurple/10 text-blurple">
              <Terminal className="w-5 h-5" />
            </div>
            <div>
              <h4 className="text-sm font-semibold text-white">Manager Bot</h4>
              <p className="text-xs text-slate-400 mt-0.5">`/bot name`, `/restart`, `/status` right in Discord.</p>
            </div>
          </div>

          <div className="p-4 rounded-xl bg-card border border-card-border flex items-start gap-3">
            <div className="p-2 rounded-lg bg-cyan-500/10 text-cyan-400">
              <Lock className="w-5 h-5" />
            </div>
            <div>
              <h4 className="text-sm font-semibold text-white">Vault Security</h4>
              <p className="text-xs text-slate-400 mt-0.5">Encrypted token storage with automated recycling.</p>
            </div>
          </div>
        </div>
      </section>

      {/* Featured Bot Catalog Showcase */}
      <section className="space-y-8">
        <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
          <div>
            <h2 className="text-2xl font-bold text-white tracking-tight">Available Bot Templates</h2>
            <p className="text-slate-400 text-sm mt-1">
              Live catalog fetched directly from the clean-architecture Catalog Microservice.
            </p>
          </div>
          <Link href="/store" className="text-sm font-semibold text-blurple hover:underline flex items-center gap-1">
            Browse All Plans <ArrowRight className="w-4 h-4" />
          </Link>
        </div>

        {loading ? (
          <div className="py-16 text-center text-slate-400 flex flex-col items-center justify-center gap-3">
            <Loader2 className="w-8 h-8 animate-spin text-blurple" />
            <p className="text-sm">Connecting to catalog-svc and loading templates...</p>
          </div>
        ) : error ? (
          <div className="p-6 rounded-2xl bg-amber-500/10 border border-amber-500/30 text-amber-300 text-sm text-center">
            {error}
          </div>
        ) : bots.length === 0 ? (
          <div className="py-12 text-center text-slate-400 text-sm">
            No bot templates found in the catalog database.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {bots.map((bot) => {
              const isMusic = bot.category === 'music';
              const isMod = bot.category === 'moderation';
              const isGame = bot.category === 'game';

              const Icon = isMusic ? Music : isMod ? Shield : isGame ? Gamepad2 : Cpu;
              const colorClass = isMusic ? 'text-cyan-400' : isMod ? 'text-amber-400' : 'text-purple-400';
              const bgClass = isMusic ? 'bg-cyan-500/10' : isMod ? 'bg-amber-500/10' : 'bg-purple-500/10';

              const plans = bot.plans || [];
              const proPlan = plans.find((p) => p.is_dedicated) || plans[0];
              const priceDisplay = proPlan
                ? proPlan.price_cents === 0
                  ? 'Free'
                  : `$${(proPlan.price_cents / 100).toFixed(2)}/mo`
                : 'Custom';

              return (
                <div
                  key={bot.id}
                  className="rounded-2xl bg-card border border-card-border p-6 flex flex-col justify-between hover:border-slate-700 transition-all group"
                >
                  <div className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div className={`p-3 rounded-xl ${bgClass} ${colorClass}`}>
                        <Icon className="w-6 h-6" />
                      </div>
                      <span className="text-[11px] font-mono px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 uppercase">
                        {bot.category}
                      </span>
                    </div>

                    <div>
                      <h3 className="text-lg font-bold text-white group-hover:text-blurple transition-colors">
                        {bot.name}
                      </h3>
                      <p className="text-xs text-slate-400 mt-1 line-clamp-3 leading-relaxed">
                        {bot.description}
                      </p>
                    </div>

                    <div className="space-y-2 pt-2 border-t border-slate-800/80">
                      <div className="text-xs text-slate-300 font-medium">Included Features:</div>
                      <ul className="space-y-1.5">
                        {proPlan?.features?.slice(0, 3).map((f, i) => (
                          <li key={i} className="text-xs text-slate-400 flex items-center gap-2">
                            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                            <span>{f}</span>
                          </li>
                        ))}
                      </ul>
                    </div>
                  </div>

                  <div className="pt-6 mt-6 border-t border-slate-800 flex items-center justify-between">
                    <div>
                      <span className="text-xs text-slate-400 block font-medium">Starting from</span>
                      <span className="text-lg font-extrabold text-white">{priceDisplay}</span>
                    </div>

                    <Link
                      href={`/store?bot=${bot.id}`}
                      className="px-4 py-2 rounded-xl bg-blurple hover:bg-blurple-hover text-white text-xs font-semibold flex items-center gap-1.5 transition-all shadow-md shadow-blurple/20"
                    >
                      Configure <ArrowRight className="w-3.5 h-3.5" />
                    </Link>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </section>

      {/* Fleet Tenancy Explanation */}
      <section className="p-8 sm:p-12 rounded-3xl bg-gradient-to-b from-card to-slate-900/60 border border-card-border space-y-8">
        <div className="max-w-3xl space-y-4">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/15 text-emerald-400 text-xs font-semibold uppercase tracking-wider">
            <Layers className="w-3.5 h-3.5" /> Multi-Subscription Fleet Architecture
          </div>
          <h2 className="text-3xl font-extrabold text-white tracking-tight">
            Run Multiple Bots & Multiple Instances in a Single Discord Guild
          </h2>
          <p className="text-slate-400 text-sm leading-relaxed">
            Need a music bot for your Main Stage and another for your VIP Lounge? Want an Aegis anti-raid shield running concurrently? Our orchestrator supports multiple active subscriptions and distinct container pods per guild with zero naming collisions.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pt-4">
          <div className="p-5 rounded-2xl bg-slate-900/80 border border-slate-800 space-y-3">
            <div className="w-8 h-8 rounded-lg bg-blurple/20 text-blurple flex items-center justify-center font-bold text-sm">
              1
            </div>
            <h4 className="text-sm font-bold text-white">Instance Labeling</h4>
            <p className="text-xs text-slate-400 leading-relaxed">
              Tag deployments with custom labels like <code>"Lobby DJ"</code> or <code>"VIP Music"</code> for seamless Discord command routing.
            </p>
          </div>

          <div className="p-5 rounded-2xl bg-slate-900/80 border border-slate-800 space-y-3">
            <div className="w-8 h-8 rounded-lg bg-blurple/20 text-blurple flex items-center justify-center font-bold text-sm">
              2
            </div>
            <h4 className="text-sm font-bold text-white">Discord Interactive Menus</h4>
            <p className="text-xs text-slate-400 leading-relaxed">
              When issuing <code>/status</code>, <code>/restart</code>, or <code>/bot name</code>, the Manager Bot provides interactive dropdowns to disambiguate bots.
            </p>
          </div>

          <div className="p-5 rounded-2xl bg-slate-900/80 border border-slate-800 space-y-3">
            <div className="w-8 h-8 rounded-lg bg-blurple/20 text-blurple flex items-center justify-center font-bold text-sm">
              3
            </div>
            <h4 className="text-sm font-bold text-white">Tabbed Fleet Dashboard</h4>
            <p className="text-xs text-slate-400 leading-relaxed">
              Inspect real-time health telemetry, latency, and memory consumption for every single instance running inside your community.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
