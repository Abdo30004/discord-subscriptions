import React from 'react';
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
} from 'lucide-react';
import { getCatalogBotsServer } from '@/lib/server-api';
import { LandingHeroActions } from '@/components/landing/LandingHeroActions';
import { BotTemplate } from '@/lib/types';

export const revalidate = 60;

export default async function HomePage() {
  const bots: BotTemplate[] = await getCatalogBotsServer();

  // Fallback default templates if catalog service is in cold-start or empty
  const displayBots: BotTemplate[] =
    bots.length > 0
      ? bots
      : [
          {
            id: 'music-template',
            name: 'Aura Music Bot',
            slug: 'aura-music',
            description: 'Ultra-low latency lossless audio streaming bot with YouTube, Spotify, and Soundcloud queue management.',
            category: 'music',
            docker_image: 'ghcr.io/discord-subscriptions/music-bot',
            default_image_tag: 'latest',
            supports_dedicated: true,
            supports_shared: true,
            plans: [
              {
                id: 'plan-music-pro',
                bot_id: 'music-template',
                name: 'Pro Audio Tier',
                description: '320kbps dedicated stream with 24/7 stage channel connectivity',
                price_cents: 799,
                currency: 'USD',
                interval: 'monthly',
                is_dedicated: true,
                features: ['320kbps Lossless Audio', '24/7 Channel Keepalive', 'Zero-Setup Turnkey Delivery', 'Kubernetes Pod Isolation'],
              },
            ],
          },
          {
            id: 'moderation-template',
            name: 'Vigilant Mod Bot',
            slug: 'vigilant-mod',
            description: 'AI-assisted anti-raid, phishing link quarantine, and automated server verification bot with raid mode protection.',
            category: 'moderation',
            docker_image: 'ghcr.io/discord-subscriptions/mod-bot',
            default_image_tag: 'latest',
            supports_dedicated: true,
            supports_shared: true,
            plans: [
              {
                id: 'plan-mod-pro',
                bot_id: 'moderation-template',
                name: 'Guardian Tier',
                description: 'Full audit logging with sub-millisecond automated raid mitigation',
                price_cents: 999,
                currency: 'USD',
                interval: 'monthly',
                is_dedicated: true,
                features: ['Anti-Phishing & Anti-Raid', 'Sub-millisecond Spam Shield', 'Granular Audit Logging', 'Vault Secret Security'],
              },
            ],
          },
          {
            id: 'rpg-template',
            name: 'Chronicles RPG Bot',
            slug: 'chronicles-rpg',
            description: 'Persistent server-wide roleplaying game with dungeon crawls, guild battles, and character leveling.',
            category: 'game',
            docker_image: 'ghcr.io/discord-subscriptions/rpg-bot',
            default_image_tag: 'latest',
            supports_dedicated: true,
            supports_shared: true,
            plans: [
              {
                id: 'plan-rpg-pro',
                bot_id: 'rpg-template',
                name: 'Guild Master Tier',
                description: 'Server-wide world events with custom boss spawns and persistent PostgreSQL saves',
                price_cents: 1299,
                currency: 'USD',
                interval: 'monthly',
                is_dedicated: true,
                features: ['Persistent PostgreSQL Saves', 'Custom Boss Encounters', 'Multi-Instance Fleet Support', 'Real-time Telemetry'],
              },
            ],
          },
        ];

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

        {/* Client-Interactive Hero Actions (Auth-Aware) */}
        <LandingHeroActions />

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

      {/* Featured Bot Catalog Showcase (Rendered on Server) */}
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

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {displayBots.map((bot) => {
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
                className="rounded-2xl bg-card border border-card-border p-6 flex flex-col justify-between hover:border-slate-700 transition-all group shadow-sm hover:shadow-md"
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
                    <p className="text-xs text-slate-400 mt-1 line-clamp-2 leading-relaxed">
                      {bot.description}
                    </p>
                  </div>

                  {proPlan && proPlan.features && (
                    <ul className="space-y-2 pt-2 border-t border-slate-800/60">
                      {proPlan.features.slice(0, 3).map((feat, idx) => (
                        <li key={idx} className="text-xs text-slate-300 flex items-center gap-2">
                          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                          <span>{feat}</span>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>

                <div className="pt-6 mt-6 border-t border-slate-800/80 flex items-center justify-between">
                  <div>
                    <div className="text-[11px] text-slate-500 uppercase tracking-wider font-semibold">
                      Starting at
                    </div>
                    <div className="text-lg font-bold text-white">{priceDisplay}</div>
                  </div>

                  <Link
                    href={`/store?bot=${bot.slug || bot.id}`}
                    className="px-4 py-2 rounded-xl bg-blurple hover:bg-blurple-hover text-white text-xs font-semibold flex items-center gap-1.5 transition-all shadow-md shadow-blurple/20"
                  >
                    Select Bot <ArrowRight className="w-3.5 h-3.5" />
                  </Link>
                </div>
              </div>
            );
          })}
        </div>
      </section>

      {/* Fleet Tenancy & Architecture Breakdown */}
      <section className="rounded-3xl bg-slate-900/60 border border-slate-800/80 p-8 sm:p-12 relative overflow-hidden">
        <div className="max-w-3xl space-y-6">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-medium">
            <Activity className="w-3.5 h-3.5" /> Production Multi-Tenancy Architecture
          </div>

          <h2 className="text-3xl font-extrabold text-white tracking-tight">
            One Guild. Multiple Bots. Zero Conflicts.
          </h2>

          <p className="text-slate-300 text-sm sm:text-base leading-relaxed">
            Need multiple music bots for different voice channels? Our platform uses label-based fleet tenancy (`instance_label`). Run a Lobby Beats bot, a VIP Lounge bot, and an Admin Mod bot simultaneously in the same Discord server without token conflicts.
          </p>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-4">
            <div className="p-4 rounded-xl bg-slate-950/60 border border-slate-800">
              <div className="font-semibold text-white text-sm flex items-center gap-2">
                <Layers className="w-4 h-4 text-cyan-400" /> Multi-Instance Fleet Isolation
              </div>
              <p className="text-xs text-slate-400 mt-1">
                Every bot deployment receives an independent Kubernetes Pod with dedicated resource bounds and automated recovery.
              </p>
            </div>

            <div className="p-4 rounded-xl bg-slate-950/60 border border-slate-800">
              <div className="font-semibold text-white text-sm flex items-center gap-2">
                <Terminal className="w-4 h-4 text-blurple" /> Interactive Discord Disambiguation
              </div>
              <p className="text-xs text-slate-400 mt-1">
                Discord slash commands (`/status`, `/restart`, `/bot name`) automatically present dropdown menus when multiple bots exist in your guild.
              </p>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
