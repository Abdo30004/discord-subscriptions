'use client';

import React, { useState, useEffect, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import {
  Server,
  Activity,
  RotateCw,
  ExternalLink,
  ShieldCheck,
  CheckCircle2,
  AlertCircle,
  Zap,
  Sparkles,
  Bot,
  Sliders,
  Clock,
  Cpu,
  Layers,
} from 'lucide-react';
import {
  MOCK_GUILDS,
  getGuildSubscriptions,
  getGuildDeployments,
  getGuildMonitoringTargets,
  restartDeployment,
  customizeBotAppearance,
} from '@/lib/api';
import { Subscription, Deployment, MonitoringTarget, DiscordGuild } from '@/lib/types';

export default function DashboardPage() {
  return (
    <Suspense fallback={<div className="max-w-7xl mx-auto py-12 px-4 text-center text-slate-400 text-sm">Loading Control Center...</div>}>
      <DashboardContent />
    </Suspense>
  );
}

function DashboardContent() {
  const searchParams = useSearchParams();
  const activatedSubId = searchParams.get('activated');

  const [selectedGuild, setSelectedGuild] = useState<DiscordGuild>(MOCK_GUILDS[0]);
  const [subscriptions, setSubscriptions] = useState<Subscription[]>([]);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [monitoringTargets, setMonitoringTargets] = useState<MonitoringTarget[]>([]);
  const [activeBotId, setActiveBotId] = useState<string>('');

  const [restarting, setRestarting] = useState(false);
  const [customizing, setCustomizing] = useState(false);
  const [customName, setCustomName] = useState('');
  const [customAvatar, setCustomAvatar] = useState('');
  const [message, setMessage] = useState<string | null>(
    activatedSubId ? 'Subscription activated! Pod is provisioned and ready.' : null
  );
  const [error, setError] = useState<string | null>(null);

  // Load guild data
  useEffect(() => {
    Promise.all([
      getGuildSubscriptions(selectedGuild.id),
      getGuildDeployments(selectedGuild.id),
      getGuildMonitoringTargets(selectedGuild.id),
    ]).then(([subs, deps, mons]) => {
      // If none returned (e.g. fresh guild in demo/dev mode), fallback to rich multi-bot demo fleet:
      if (deps.length === 0 && subs.length === 0) {
        const demoSub1: Subscription = {
          id: 'sub-demo-music-1',
          user_id: 'user-discord-007',
          guild_id: selectedGuild.id,
          plan_id: 'plan-music-pro',
          bot_type: 'music',
          instance_label: 'Main Stage',
          status: 'active',
          provider: 'paypal',
          is_dedicated: true,
          is_zero_setup: true,
          valid_until: new Date(Date.now() + 28 * 24 * 3600 * 1000).toISOString(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        const demoSub2: Subscription = {
          id: 'sub-demo-mod-2',
          user_id: 'user-discord-007',
          guild_id: selectedGuild.id,
          plan_id: 'plan-mod-pro',
          bot_type: 'moderation',
          instance_label: 'Security Aegis',
          status: 'active',
          provider: 'paypal',
          is_dedicated: true,
          is_zero_setup: true,
          valid_until: new Date(Date.now() + 30 * 24 * 3600 * 1000).toISOString(),
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        const demoDep1: Deployment = {
          id: 'dep-music-cluster-1',
          subscription_id: 'sub-demo-music-1',
          user_id: 'user-discord-007',
          guild_id: selectedGuild.id,
          bot_type: 'music',
          instance_label: 'Main Stage',
          k8s_namespace: 'discord-bots',
          k8s_deployment_name: `bot-${selectedGuild.id}-m1`,
          image_name: 'ghcr.io/discord-subscriptions/music-bot',
          image_tag: 'v1.2.0',
          status: 'running',
          is_zero_setup: true,
          client_id: '131234567890123456',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        const demoDep2: Deployment = {
          id: 'dep-mod-cluster-2',
          subscription_id: 'sub-demo-mod-2',
          user_id: 'user-discord-007',
          guild_id: selectedGuild.id,
          bot_type: 'moderation',
          instance_label: 'Security Aegis',
          k8s_namespace: 'discord-bots',
          k8s_deployment_name: `bot-${selectedGuild.id}-mod2`,
          image_name: 'ghcr.io/discord-subscriptions/mod-bot',
          image_tag: 'v2.0.1',
          status: 'running',
          is_zero_setup: true,
          client_id: '131234567890123457',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        setSubscriptions([demoSub1, demoSub2]);
        setDeployments([demoDep1, demoDep2]);
        setActiveBotId(demoDep1.id);
        setMonitoringTargets([
          {
            id: 'tgt-1',
            bot_id: demoDep1.id,
            guild_id: selectedGuild.id,
            instance_label: 'Main Stage',
            health_url: 'http://bot1/health',
            is_active: true,
            current_status: 'online',
            consecutive_failures: 0,
            recent_logs: [
              {
                id: 1,
                target_id: 'tgt-1',
                status: 'online',
                status_code: 200,
                latency_ms: 22,
                discord_ping_ms: 18,
                memory_usage_mb: 138,
                checked_at: new Date().toISOString(),
              },
            ],
          },
          {
            id: 'tgt-2',
            bot_id: demoDep2.id,
            guild_id: selectedGuild.id,
            instance_label: 'Security Aegis',
            health_url: 'http://bot2/health',
            is_active: true,
            current_status: 'online',
            consecutive_failures: 0,
            recent_logs: [
              {
                id: 2,
                target_id: 'tgt-2',
                status: 'online',
                status_code: 200,
                latency_ms: 14,
                discord_ping_ms: 12,
                memory_usage_mb: 85,
                checked_at: new Date().toISOString(),
              },
            ],
          },
        ]);
      } else {
        setSubscriptions(subs);
        setDeployments(deps);
        setMonitoringTargets(mons);
        if (deps.length > 0) {
          setActiveBotId(deps[0].id);
        } else if (subs.length > 0) {
          setActiveBotId(subs[0].id);
        }
      }
    });
  }, [selectedGuild]);

  const currentDeployment = deployments.find((d) => d.id === activeBotId) || deployments[0] || null;
  const currentSubscription = subscriptions.find((s) => s.id === currentDeployment?.subscription_id) || subscriptions[0] || null;
  const currentMonitoring = monitoringTargets.find((m) => m.bot_id === currentDeployment?.id) || monitoringTargets[0] || null;

  const handleRestart = async () => {
    if (!currentDeployment) return;
    setRestarting(true);
    setMessage(null);
    setError(null);
    try {
      await restartDeployment(currentDeployment.id);
      setMessage(`Rolling restart initiated for ${currentDeployment.instance_label || currentDeployment.bot_type}. Reconnection in ~15s.`);
    } catch (err: any) {
      setError(err.message || 'Failed to restart bot');
    } finally {
      setRestarting(false);
    }
  };

  const handleCustomize = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!customName && !customAvatar) return;
    if (!currentDeployment) return;
    setCustomizing(true);
    setMessage(null);
    setError(null);
    try {
      await customizeBotAppearance(currentDeployment.id, customName || undefined, customAvatar || undefined);
      setMessage(`Bot appearance for "${currentDeployment.instance_label || currentDeployment.bot_type}" successfully updated! Discord rate limits: max 2 updates per hour.`);
      setCustomName('');
      setCustomAvatar('');
    } catch (err: any) {
      setError(err.message || 'Failed to customize bot');
    } finally {
      setCustomizing(false);
    }
  };

  const botInviteUrl = currentDeployment?.client_id
    ? `https://discord.com/api/oauth2/authorize?client_id=${currentDeployment.client_id}&permissions=8&scope=bot%20applications.commands`
    : `https://discord.com/api/oauth2/authorize?client_id=131234567890123456&permissions=8&scope=bot%20applications.commands`;

  return (
    <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8 space-y-8">
      {/* Top Banner / Guild Selector */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-6 border-b border-card-border">
        <div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">Bot Operations Control Center</h1>
          <p className="text-slate-400 text-sm mt-1">Manage multi-bot subscriptions, dedicated containers, and per-instance personas.</p>
        </div>

        {/* Guild Selector */}
        <div className="flex items-center gap-3">
          <span className="text-xs font-mono text-slate-400">Current Server:</span>
          <select
            value={selectedGuild.id}
            onChange={(e) => {
              const g = MOCK_GUILDS.find((x) => x.id === e.target.value);
              if (g) setSelectedGuild(g);
            }}
            className="bg-card border border-slate-700 text-white rounded-xl px-4 py-2.5 text-xs font-semibold focus:outline-none focus:border-blurple"
          >
            {MOCK_GUILDS.map((g) => (
              <option key={g.id} value={g.id}>
                {g.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Bot Fleet Selector Tabs */}
      {deployments.length > 0 && (
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-2xl bg-card border border-card-border">
          <div className="flex items-center gap-2 overflow-x-auto pb-1 sm:pb-0">
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-800 text-xs font-mono text-slate-400 shrink-0">
              <Layers className="w-3.5 h-3.5 text-blurple" /> Bot Fleet ({deployments.length}):
            </div>
            {deployments.map((d, idx) => {
              const isSelected = currentDeployment?.id === d.id;
              const target = monitoringTargets.find((t) => t.bot_id === d.id);
              const status = target?.current_status || d.status || 'running';
              return (
                <button
                  key={d.id}
                  type="button"
                  onClick={() => setActiveBotId(d.id)}
                  className={`flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-semibold whitespace-nowrap transition-all ${
                    isSelected
                      ? 'bg-blurple text-white shadow-md shadow-blurple/25 border border-blurple'
                      : 'bg-slate-900/80 hover:bg-slate-800 text-slate-300 border border-slate-800'
                  }`}
                >
                  <span
                    className={`w-2 h-2 rounded-full ${
                      status === 'online' ? 'bg-emerald-400' : status === 'degraded' ? 'bg-amber-400' : 'bg-red-400'
                    }`}
                  />
                  <span>{d.instance_label || `${d.bot_type} #${idx + 1}`}</span>
                  <span
                    className={`text-[10px] uppercase font-mono px-1.5 py-0.5 rounded ${
                      isSelected ? 'bg-white/20' : 'bg-slate-800 text-slate-400'
                    }`}
                  >
                    {d.bot_type}
                  </span>
                </button>
              );
            })}
          </div>

          <Link
            href="/store"
            className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-white text-xs font-semibold border border-slate-700 transition-colors shrink-0"
          >
            <Sparkles className="w-3.5 h-3.5 text-amber-400" />
            <span>+ Deploy Another Bot</span>
          </Link>
        </div>
      )}

      {message && (
        <div className="p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-medium flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4 shrink-0" />
          <span>{message}</span>
        </div>
      )}

      {error && (
        <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-xs font-medium flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Grid: Status Cards & Metrics for Selected Instance */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Subscription Status Card */}
        <div className="glass-card rounded-2xl p-6 space-y-4 border border-card-border">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Active License</span>
            <span className="px-2.5 py-0.5 rounded-full text-[11px] font-mono font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 uppercase">
              {currentSubscription?.status || 'Active'}
            </span>
          </div>

          <div>
            <div className="flex items-center gap-2">
              <h3 className="text-lg font-bold text-white capitalize">
                {currentDeployment?.instance_label || currentSubscription?.bot_type || 'Bot'}
              </h3>
              <span className="text-xs font-mono uppercase bg-slate-800 text-slate-400 px-2 py-0.5 rounded">
                {currentSubscription?.bot_type || currentDeployment?.bot_type || 'Music'}
              </span>
            </div>
            <p className="text-xs text-slate-400 mt-1">Single-tenant isolated cluster container.</p>
          </div>

          <div className="pt-4 border-t border-slate-800/80 space-y-2 text-xs">
            <div className="flex justify-between text-slate-400">
              <span>Billing Provider:</span>
              <span className="text-white capitalize font-medium">{currentSubscription?.provider || 'PayPal'}</span>
            </div>
            <div className="flex justify-between text-slate-400">
              <span>Deployment Mode:</span>
              <span className="text-amber-400 font-medium flex items-center gap-1">
                {currentSubscription?.is_zero_setup ? (
                  <>
                    <Zap className="w-3.5 h-3.5" /> Turnkey Managed
                  </>
                ) : (
                  'Self-Hosted Token'
                )}
              </span>
            </div>
            <div className="flex justify-between text-slate-400">
              <span>Renews On:</span>
              <span className="text-white font-mono">
                {currentSubscription ? new Date(currentSubscription.valid_until).toLocaleDateString() : 'N/A'}
              </span>
            </div>
          </div>
        </div>

        {/* Kubernetes Pod Health & Telemetry */}
        <div className="glass-card rounded-2xl p-6 space-y-4 border border-card-border">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Pod Health & Ping</span>
            <span className="flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[11px] font-mono font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
              {currentMonitoring?.current_status || 'Online'}
            </span>
          </div>

          <div className="grid grid-cols-2 gap-3 pt-1">
            <div className="p-3 rounded-xl bg-slate-900/60 border border-slate-800">
              <span className="text-[11px] text-slate-400">Discord Ping</span>
              <p className="text-lg font-bold text-white font-mono mt-0.5">
                {currentMonitoring?.recent_logs?.[0]?.discord_ping_ms || 18}
                <span className="text-xs font-normal text-slate-400">ms</span>
              </p>
            </div>
            <div className="p-3 rounded-xl bg-slate-900/60 border border-slate-800">
              <span className="text-[11px] text-slate-400">Memory Usage</span>
              <p className="text-lg font-bold text-white font-mono mt-0.5">
                {currentMonitoring?.recent_logs?.[0]?.memory_usage_mb || 142}
                <span className="text-xs font-normal text-slate-400">MB</span>
              </p>
            </div>
          </div>

          <div className="pt-2 flex items-center justify-between">
            <span className="text-[11px] text-slate-400 font-mono">
              Pod: {currentDeployment?.k8s_deployment_name || `bot-${selectedGuild.id}`}
            </span>
            <button
              type="button"
              onClick={handleRestart}
              disabled={restarting}
              className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-white text-xs font-semibold flex items-center gap-1.5 border border-slate-700 transition-colors"
            >
              <RotateCw className={`w-3.5 h-3.5 ${restarting ? 'animate-spin' : ''}`} />
              Restart Container
            </button>
          </div>
        </div>

        {/* Discord Bot Direct Invite & Quick Links */}
        <div className="glass-card rounded-2xl p-6 space-y-4 border border-card-border flex flex-col justify-between">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Bot Installation</span>
              <span className="text-[10px] px-2 py-0.5 rounded bg-blurple/20 text-blurple font-mono">
                OAuth2 Ready
              </span>
            </div>

            <div>
              <h4 className="text-sm font-bold text-white">Join Bot to Your Server</h4>
              <p className="text-xs text-slate-400 mt-1">
                Authorizes <strong className="text-white">{currentDeployment?.instance_label || currentDeployment?.bot_type || 'Bot'}</strong> to join your Discord server with 1 click.
              </p>
            </div>
          </div>

          <div className="space-y-2 pt-4">
            <a
              href={botInviteUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="w-full py-2.5 px-4 rounded-xl bg-blurple hover:bg-blurple-hover text-white text-xs font-bold flex items-center justify-center gap-2 shadow-lg shadow-blurple/25 transition-all"
            >
              <Bot className="w-4 h-4" /> Add Bot to Server <ExternalLink className="w-3.5 h-3.5" />
            </a>

            {!currentSubscription?.is_zero_setup && (
              <Link
                href={`/setup/${currentSubscription?.id || 'manual'}`}
                className="w-full py-2 px-4 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium flex items-center justify-center gap-1.5 transition-colors"
              >
                Configure Custom Bot Token
              </Link>
            )}
          </div>
        </div>
      </div>

      {/* Bot Persona Customization Section */}
      <div className="p-6 rounded-2xl bg-card border border-card-border space-y-6">
        <div>
          <h3 className="text-base font-bold text-white flex items-center gap-2">
            <Sliders className="w-4 h-4 text-blurple" /> In-Discord Appearance Customization: {currentDeployment?.instance_label || currentDeployment?.bot_type || 'Bot'}
          </h3>
          <p className="text-xs text-slate-400 mt-1">
            Change the name and profile picture for this bot instance. You can also customize anytime inside Discord using{' '}
            <code className="text-blurple font-mono bg-slate-900 px-1 py-0.5 rounded">/bot name</code> and{' '}
            <code className="text-blurple font-mono bg-slate-900 px-1 py-0.5 rounded">/bot avatar</code>.
          </p>
        </div>

        <form onSubmit={handleCustomize} className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="text-xs text-slate-400 block mb-1">New Bot Username</label>
            <input
              type="text"
              value={customName}
              onChange={(e) => setCustomName(e.target.value)}
              placeholder="e.g. VIP Beats 24/7"
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3.5 py-2 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blurple"
            />
          </div>

          <div>
            <label className="text-xs text-slate-400 block mb-1">New Avatar Image URL</label>
            <input
              type="url"
              value={customAvatar}
              onChange={(e) => setCustomAvatar(e.target.value)}
              placeholder="https://cdn.discordapp.com/... or direct image link"
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3.5 py-2 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blurple"
            />
          </div>

          <div className="flex items-end">
            <button
              type="submit"
              disabled={customizing || (!customName && !customAvatar)}
              className="w-full py-2.5 px-4 rounded-lg bg-blurple hover:bg-blurple-hover disabled:bg-slate-800 disabled:text-slate-500 text-white text-xs font-semibold transition-colors"
            >
              {customizing ? 'Applying to Discord API...' : 'Update Bot Appearance'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
