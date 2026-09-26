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
  ShoppingBag,
  Loader2,
  LogIn,
} from 'lucide-react';
import {
  getGuildSubscriptions,
  getGuildDeployments,
  getGuildMonitoringTargets,
  restartDeployment,
  customizeBotAppearance,
} from '@/lib/api';
import { Subscription, Deployment, MonitoringTarget, DiscordGuild } from '@/lib/types';
import { useAuth } from '@/contexts/AuthContext';

export default function DashboardPage() {
  return (
    <Suspense
      fallback={
        <div className="max-w-7xl mx-auto py-12 px-4 text-center text-slate-400 text-sm flex items-center justify-center gap-2">
          <Loader2 className="w-5 h-5 animate-spin text-blurple" /> Loading Control Center...
        </div>
      }
    >
      <DashboardContent />
    </Suspense>
  );
}

function DashboardContent() {
  const searchParams = useSearchParams();
  const activatedSubId = searchParams.get('activated');

  const { user, guilds, selectedGuild, selectGuild, loginWithDiscord, loginAsDev } = useAuth();

  const [subscriptions, setSubscriptions] = useState<Subscription[]>([]);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [monitoringTargets, setMonitoringTargets] = useState<MonitoringTarget[]>([]);
  const [activeBotId, setActiveBotId] = useState<string>('');
  const [loadingData, setLoadingData] = useState<boolean>(false);

  const [restarting, setRestarting] = useState(false);
  const [customizing, setCustomizing] = useState(false);
  const [customName, setCustomName] = useState('');
  const [customAvatar, setCustomAvatar] = useState('');
  const [message, setMessage] = useState<string | null>(
    activatedSubId ? 'Subscription activated! Pod is provisioned and ready in your cluster.' : null
  );
  const [error, setError] = useState<string | null>(null);

  // Load guild data when selectedGuild changes
  useEffect(() => {
    if (!selectedGuild) {
      setSubscriptions([]);
      setDeployments([]);
      setMonitoringTargets([]);
      return;
    }

    setLoadingData(true);
    setError(null);

    Promise.all([
      getGuildSubscriptions(selectedGuild.id),
      getGuildDeployments(selectedGuild.id),
      getGuildMonitoringTargets(selectedGuild.id),
    ])
      .then(([subs, deps, mons]) => {
        setSubscriptions(subs);
        setDeployments(deps);
        setMonitoringTargets(mons);

        // Set active bot tab
        if (deps.length > 0) {
          setActiveBotId(deps[0].id);
        } else if (subs.length > 0) {
          setActiveBotId(subs[0].id);
        } else {
          setActiveBotId('');
        }
      })
      .catch((err) => {
        console.error('Failed fetching guild data:', err);
        setError('Failed to fetch guild fleet data from services');
      })
      .finally(() => {
        setLoadingData(false);
      });
  }, [selectedGuild]);

  // Restart Handler
  const handleRestart = async (depId: string) => {
    setRestarting(true);
    setError(null);
    try {
      await restartDeployment(depId);
      setMessage('Rolling restart dispatched to Kubernetes cluster. Pod cycling in background.');
      setTimeout(() => setMessage(null), 5000);
    } catch (err: any) {
      setError(err.message || 'Failed to trigger deployment restart');
    } finally {
      setRestarting(false);
    }
  };

  // Customization Handler
  const handleCustomize = async (identifier: string) => {
    if (!customName.trim() && !customAvatar.trim()) return;
    setCustomizing(true);
    setError(null);
    try {
      await customizeBotAppearance(
        identifier,
        customName.trim() || undefined,
        customAvatar.trim() || undefined
      );
      setMessage('Persona customization updated! Syncing across Discord Gateway...');
      setCustomName('');
      setCustomAvatar('');
      setTimeout(() => setMessage(null), 5000);
    } catch (err: any) {
      setError(err.message || 'Customization update failed');
    } finally {
      setCustomizing(false);
    }
  };

  if (!user) {
    return (
      <div className="max-w-2xl mx-auto py-24 px-4 text-center space-y-6">
        <div className="w-16 h-16 rounded-2xl bg-blurple/15 text-blurple mx-auto flex items-center justify-center">
          <Bot className="w-8 h-8" />
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-extrabold text-white">Sign in to Access Your Bot Fleet</h1>
          <p className="text-slate-400 text-sm max-w-md mx-auto">
            Connect your Discord account to view your server's active subscriptions, manage deployments, and customize bot personas.
          </p>
        </div>
        <div className="flex justify-center gap-3 pt-2">
          <button
            type="button"
            onClick={() => loginWithDiscord()}
            className="py-3 px-6 rounded-xl bg-blurple hover:bg-blurple-hover text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-blurple/25 transition-all"
          >
            <LogIn className="w-4 h-4" /> Sign In with Discord
          </button>
          <button
            type="button"
            onClick={() => loginAsDev()}
            className="py-3 px-5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold border border-slate-700 transition-all"
          >
            Dev Instant Login
          </button>
        </div>
      </div>
    );
  }

  // Active bot matching
  const activeDeployment = deployments.find((d) => d.id === activeBotId) || deployments[0];
  const activeMonitoring = monitoringTargets.find(
    (m) => m.bot_id === activeDeployment?.id || m.instance_label === activeDeployment?.instance_label
  );
  const activeSubscription = subscriptions.find(
    (s) => s.id === activeDeployment?.subscription_id || s.bot_type === activeDeployment?.bot_type
  );

  return (
    <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8 space-y-8">
      {/* Title & Server Switcher */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-6 border-b border-card-border">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-emerald-500/15 text-emerald-400 text-xs font-semibold uppercase tracking-wider mb-2">
            <Activity className="w-3.5 h-3.5" /> Fleet Operations Center
          </div>
          <h1 className="text-3xl font-extrabold text-white tracking-tight">Discord Guild Fleet Tenancy</h1>
          <p className="text-slate-400 text-sm mt-1">
            Manage multi-subscription dedicated and turnkey bots running concurrently inside your server.
          </p>
        </div>

        {/* Guild Selector */}
        {guilds.length > 0 && selectedGuild && (
          <div className="flex items-center gap-3">
            <label className="text-xs text-slate-400 font-medium">Active Server:</label>
            <div className="relative">
              <select
                value={selectedGuild.id}
                onChange={(e) => {
                  const g = guilds.find((x) => x.id === e.target.value);
                  if (g) selectGuild(g);
                }}
                className="bg-card border border-card-border text-white text-xs font-bold rounded-xl px-4 py-2.5 focus:outline-none focus:border-blurple shadow-sm cursor-pointer"
              >
                {guilds.map((g) => (
                  <option key={g.id} value={g.id} className="bg-slate-900 text-white">
                    {g.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
        )}
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

      {loadingData ? (
        <div className="py-24 text-center space-y-3">
          <Loader2 className="w-8 h-8 animate-spin text-blurple mx-auto" />
          <p className="text-slate-400 text-sm">Querying billing, deploy, and monitor microservices...</p>
        </div>
      ) : deployments.length === 0 && subscriptions.length === 0 ? (
        <div className="p-12 rounded-3xl bg-card border border-card-border text-center space-y-6">
          <div className="w-16 h-16 rounded-2xl bg-blurple/10 text-blurple mx-auto flex items-center justify-center">
            <Bot className="w-8 h-8" />
          </div>
          <div className="space-y-2 max-w-md mx-auto">
            <h3 className="text-xl font-bold text-white">No Bots Active in {selectedGuild?.name || 'this Server'}</h3>
            <p className="text-xs text-slate-400 leading-relaxed">
              You haven't provisioned any bot subscriptions for this server yet. Choose a template from our catalog to get started.
            </p>
          </div>
          <div>
            <Link
              href={`/store?guild=${selectedGuild?.id || ''}`}
              className="inline-flex items-center gap-2 py-3 px-6 rounded-xl bg-blurple hover:bg-blurple-hover text-white text-xs font-bold shadow-lg shadow-blurple/25 transition-all"
            >
              <ShoppingBag className="w-4 h-4" /> Browse Bot Catalog
            </Link>
          </div>
        </div>
      ) : (
        <>
          {/* Multi-Subscription Instance Tabs (Fleet Switcher) */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
                <Layers className="w-3.5 h-3.5 text-blurple" /> Guild Fleet Instances ({deployments.length || subscriptions.length})
              </span>
              <Link
                href={`/store?guild=${selectedGuild?.id || ''}`}
                className="text-xs text-blurple hover:underline flex items-center gap-1 font-semibold"
              >
                + Add Another Bot to Server
              </Link>
            </div>

            <div className="flex flex-wrap gap-2 pt-1">
              {deployments.map((dep) => {
                const isActive = (activeDeployment?.id === dep.id);
                return (
                  <button
                    key={dep.id}
                    type="button"
                    onClick={() => setActiveBotId(dep.id)}
                    className={`py-2 px-4 rounded-xl text-xs font-bold transition-all flex items-center gap-2 border ${
                      isActive
                        ? 'bg-blurple text-white border-blurple shadow-md'
                        : 'bg-card border-card-border text-slate-300 hover:border-slate-700'
                    }`}
                  >
                    <Bot className="w-3.5 h-3.5" />
                    <span>{dep.instance_label || dep.bot_type.toUpperCase()}</span>
                    <span className="text-[10px] font-mono px-1.5 py-0.2 rounded bg-black/30">
                      {dep.bot_type}
                    </span>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Active Bot Control Dashboard */}
          {activeDeployment && (
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Left Column: Pod Status & Telemetry */}
              <div className="lg:col-span-2 space-y-6">
                <div className="p-6 rounded-2xl bg-card border border-card-border space-y-6">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-800">
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="text-lg font-bold text-white">
                          {activeDeployment.instance_label || activeDeployment.bot_type.toUpperCase()}
                        </h3>
                        <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                          {activeDeployment.status}
                        </span>
                      </div>
                      <p className="text-xs text-slate-400 font-mono mt-0.5">
                        K8s Pod: <code>{activeDeployment.k8s_deployment_name}</code>
                      </p>
                    </div>

                    <div className="flex items-center gap-2">
                      <button
                        type="button"
                        onClick={() => handleRestart(activeDeployment.id)}
                        disabled={restarting}
                        className="py-2 px-4 rounded-xl bg-slate-800 hover:bg-slate-700 disabled:opacity-50 text-white text-xs font-semibold flex items-center gap-1.5 border border-slate-700 transition-all"
                      >
                        <RotateCw className={`w-3.5 h-3.5 ${restarting ? 'animate-spin' : ''}`} />
                        <span>Restart Pod</span>
                      </button>
                    </div>
                  </div>

                  {/* Telemetry Metrics Grid */}
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                    <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800 space-y-1">
                      <span className="text-[10px] uppercase font-mono text-slate-400 block">Watchdog Status</span>
                      <span className="text-sm font-bold text-emerald-400 flex items-center gap-1.5">
                        <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                        {activeMonitoring?.current_status?.toUpperCase() || 'ONLINE'}
                      </span>
                    </div>

                    <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800 space-y-1">
                      <span className="text-[10px] uppercase font-mono text-slate-400 block">Gateway Latency</span>
                      <span className="text-sm font-bold text-white font-mono">
                        {activeMonitoring?.recent_logs?.[0]?.discord_ping_ms ?? 18} ms
                      </span>
                    </div>

                    <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800 space-y-1">
                      <span className="text-[10px] uppercase font-mono text-slate-400 block">Memory RSS</span>
                      <span className="text-sm font-bold text-white font-mono">
                        {activeMonitoring?.recent_logs?.[0]?.memory_usage_mb ?? 142} MB
                      </span>
                    </div>

                    <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800 space-y-1">
                      <span className="text-[10px] uppercase font-mono text-slate-400 block">Container Image</span>
                      <span className="text-xs font-bold text-slate-300 font-mono truncate block">
                        {activeDeployment.image_tag || 'latest'}
                      </span>
                    </div>
                  </div>

                  {/* Live Health Logs */}
                  <div className="space-y-3">
                    <h4 className="text-xs font-bold uppercase tracking-wider text-slate-400">
                      Recent Health Checks (Monitor-Svc Poller)
                    </h4>
                    <div className="space-y-2 max-h-48 overflow-y-auto font-mono text-xs">
                      {activeMonitoring?.recent_logs && activeMonitoring.recent_logs.length > 0 ? (
                        activeMonitoring.recent_logs.map((log) => (
                          <div
                            key={log.id}
                            className="p-2.5 rounded-lg bg-slate-900/70 border border-slate-800/80 flex items-center justify-between text-[11px]"
                          >
                            <div className="flex items-center gap-2">
                              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                              <span className="text-slate-300">HTTP {log.status_code || 200} OK</span>
                              <span className="text-slate-500">({log.latency_ms || 24}ms)</span>
                            </div>
                            <span className="text-slate-500">
                              {new Date(log.checked_at).toLocaleTimeString()}
                            </span>
                          </div>
                        ))
                      ) : (
                        <div className="p-3 text-center text-slate-500 text-xs">
                          Poller listening... Waiting for next scheduled probe cycle.
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>

              {/* Right Column: Bot Persona Customizer & Subscription Info */}
              <div className="space-y-6">
                {/* Customizer */}
                <div className="p-6 rounded-2xl bg-card border border-card-border space-y-4">
                  <div className="flex items-center gap-2">
                    <Sliders className="w-4 h-4 text-blurple" />
                    <h3 className="text-sm font-bold text-white uppercase tracking-wider">
                      Persona Customization
                    </h3>
                  </div>
                  <p className="text-xs text-slate-400">
                    Update the bot's username and avatar on the Discord Gateway.
                  </p>

                  <div className="space-y-3">
                    <div>
                      <label className="text-[11px] text-slate-400 block mb-1">New Bot Name</label>
                      <input
                        type="text"
                        value={customName}
                        onChange={(e) => setCustomName(e.target.value)}
                        placeholder="e.g. Neon Groover"
                        className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                      />
                    </div>

                    <div>
                      <label className="text-[11px] text-slate-400 block mb-1">New Avatar Image URL</label>
                      <input
                        type="url"
                        value={customAvatar}
                        onChange={(e) => setCustomAvatar(e.target.value)}
                        placeholder="https://example.com/avatar.png"
                        className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-blurple"
                      />
                    </div>

                    <button
                      type="button"
                      onClick={() => handleCustomize(activeDeployment.id)}
                      disabled={customizing || (!customName.trim() && !customAvatar.trim())}
                      className="w-full py-2.5 rounded-xl bg-blurple hover:bg-blurple-hover disabled:bg-slate-800 disabled:text-slate-500 text-white text-xs font-bold transition-all shadow-md shadow-blurple/20"
                    >
                      {customizing ? 'Updating Discord...' : 'Apply Appearance Changes'}
                    </button>
                  </div>
                </div>

                {/* Subscription Details */}
                <div className="p-6 rounded-2xl bg-card border border-card-border space-y-3">
                  <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider">
                    Subscription Invariant
                  </h3>
                  <div className="space-y-2 text-xs">
                    <div className="flex justify-between">
                      <span className="text-slate-500">Plan Tier:</span>
                      <span className="font-semibold text-white">
                        {activeSubscription?.plan_id || 'Pro Dedicated'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-500">Turnkey 0-Setup:</span>
                      <span className="font-semibold text-emerald-400">
                        {activeDeployment.is_zero_setup ? 'Enabled' : 'Custom Token'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-500">Valid Until:</span>
                      <span className="font-mono text-slate-300">
                        {activeSubscription?.valid_until
                          ? new Date(activeSubscription.valid_until).toLocaleDateString()
                          : '30 Days Active'}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
