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
import { MOCK_BOTS } from '@/lib/api';

export default function HomePage() {
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
            <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">Available Bot Templates</h2>
            <p className="text-sm text-slate-400 mt-1">High-performance production templates tested and audited for maximum uptime.</p>
          </div>
          <Link href="/store" className="text-sm font-medium text-blurple hover:underline flex items-center gap-1">
            View all pricing tiers <ArrowRight className="w-3.5 h-3.5" />
          </Link>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {MOCK_BOTS.map((bot) => {
            const isMusic = bot.category === 'music';
            const isMod = bot.category === 'moderation';
            return (
              <div
                key={bot.id}
                className="glass-card rounded-2xl p-6 flex flex-col justify-between relative overflow-hidden group"
              >
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div
                      className={`w-12 h-12 rounded-xl flex items-center justify-center text-white ${
                        isMusic
                          ? 'bg-gradient-to-tr from-purple-600 to-indigo-600'
                          : isMod
                          ? 'bg-gradient-to-tr from-blue-600 to-cyan-600'
                          : 'bg-gradient-to-tr from-emerald-600 to-teal-600'
                      }`}
                    >
                      {isMusic ? (
                        <Music className="w-6 h-6" />
                      ) : isMod ? (
                        <Shield className="w-6 h-6" />
                      ) : (
                        <Gamepad2 className="w-6 h-6" />
                      )}
                    </div>
                    <span className="text-xs uppercase font-mono px-2.5 py-1 rounded-md bg-slate-800 text-slate-300 border border-slate-700">
                      {bot.category}
                    </span>
                  </div>

                  <div>
                    <h3 className="text-xl font-bold text-white group-hover:text-blurple transition-colors">
                      {bot.name}
                    </h3>
                    <p className="text-xs text-slate-400 mt-2 leading-relaxed">{bot.description}</p>
                  </div>

                  {bot.plans && bot.plans[0] && (
                    <div className="pt-2 border-t border-slate-800/80 space-y-2">
                      <p className="text-xs font-semibold uppercase tracking-wider text-slate-400">Included Features</p>
                      <ul className="space-y-1.5 text-xs text-slate-300">
                        {bot.plans[bot.plans.length - 1].features.slice(0, 3).map((f, idx) => (
                          <li key={idx} className="flex items-center gap-2">
                            <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                            <span>{f}</span>
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>

                <div className="pt-6 mt-6 border-t border-slate-800 flex items-center justify-between">
                  <div>
                    <span className="text-xs text-slate-400">Starting at</span>
                    <p className="text-lg font-bold text-white">
                      ${(bot.plans?.[1]?.price_cents || bot.plans?.[0]?.price_cents || 799) / 100}
                      <span className="text-xs font-normal text-slate-400">/mo</span>
                    </p>
                  </div>
                  <Link
                    href={`/store?bot=${bot.id}`}
                    className="px-4 py-2 rounded-lg bg-blurple hover:bg-blurple-hover text-white text-xs font-semibold flex items-center gap-1.5 transition-colors"
                  >
                    Select Plan <ArrowRight className="w-3.5 h-3.5" />
                  </Link>
                </div>
              </div>
            );
          })}
        </div>
      </section>

      {/* Architecture Deep Dive Interactive Section */}
      <section className="p-8 rounded-3xl bg-card border border-card-border space-y-8">
        <div className="text-center max-w-3xl mx-auto space-y-2">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-slate-800 text-xs font-mono text-slate-300">
            <Layers className="w-3.5 h-3.5 text-blurple" /> Event-Driven Microservices Architecture
          </div>
          <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">Built for Scale and Reliability</h2>
          <p className="text-sm text-slate-400">
            Every subscription event flows through RabbitMQ message brokers, HashiCorp Vault secrets, and Kubernetes pods.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
          <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-mono font-bold text-blurple">auth-svc</span>
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400">:8080</span>
            </div>
            <p className="text-xs text-slate-400">Discord OAuth2, JWT issuance & Admin permission bitmask verification.</p>
          </div>

          <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-mono font-bold text-blurple">catalog-svc</span>
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400">:8081</span>
            </div>
            <p className="text-xs text-slate-400">Bot templates, dedicated container image tags & subscription plan tiers.</p>
          </div>

          <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-mono font-bold text-blurple">billing-svc</span>
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400">:8082</span>
            </div>
            <p className="text-xs text-slate-400">PayPal subscriptions, promo codes, gift card vouchers & admin grants.</p>
          </div>

          <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-mono font-bold text-blurple">deploy-svc</span>
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400">:8083</span>
            </div>
            <p className="text-xs text-slate-400">K8s client-go orchestration, pre-warmed token pool & Vault injection.</p>
          </div>

          <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-mono font-bold text-blurple">monitor-svc</span>
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400">:8084</span>
            </div>
            <p className="text-xs text-slate-400">Active health check pinger, WebSocket telemetry & failover notifications.</p>
          </div>
        </div>
      </section>
    </div>
  );
}
