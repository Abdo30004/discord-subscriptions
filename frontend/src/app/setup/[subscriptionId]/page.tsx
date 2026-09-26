'use client';

import React, { useState } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import {
  Key,
  Shield,
  ExternalLink,
  CheckCircle2,
  AlertCircle,
  Loader2,
  ArrowRight,
  Bot,
  Zap,
} from 'lucide-react';
import { provisionDeployment } from '@/lib/api';
import { useAuth } from '@/contexts/AuthContext';

export default function SetupWizardPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const router = useRouter();
  const { user, selectedGuild } = useAuth();

  const subId = (params.subscriptionId as string) || '';
  const guildId = searchParams.get('guild') || selectedGuild?.id || '';
  const botType = searchParams.get('botType') || 'music';

  const [botToken, setBotToken] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleSaveToken = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!botToken.trim()) {
      setError('Please provide a Discord bot token');
      return;
    }
    if (!user) {
      setError('You must be logged in to provision a bot pod');
      return;
    }

    setLoading(true);
    setError(null);
    try {
      await provisionDeployment({
        subscription_id: subId,
        user_id: user.id,
        guild_id: guildId || '0',
        bot_type: botType,
        bot_token: botToken.trim(),
        image_tag: 'latest',
        is_zero_setup: false,
      });

      setSuccess(true);
      setTimeout(() => {
        router.push('/dashboard');
      }, 1500);
    } catch (err: any) {
      setError(err.message || 'Failed to provision deployment pod in deploy-svc');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-3xl mx-auto py-12 px-4 sm:px-6 lg:px-8 space-y-8">
      {/* Header */}
      <div>
        <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-blurple/15 border border-blurple/30 text-blurple text-xs font-semibold uppercase tracking-wider mb-2">
          <Key className="w-3.5 h-3.5" /> Self-Setup Wizard
        </div>
        <h1 className="text-3xl font-extrabold text-white tracking-tight">Connect Your Custom Bot Token</h1>
        <p className="text-slate-400 text-sm mt-1">
          Follow the 3-step walkthrough below to attach your personal Discord application to this dedicated Kubernetes instance.
        </p>
      </div>

      {success && (
        <div className="p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-medium flex items-center gap-2">
          <CheckCircle2 className="w-5 h-5 shrink-0" />
          <span>Token saved securely in HashiCorp Vault! Starting pod container and redirecting to Dashboard...</span>
        </div>
      )}

      {error && (
        <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-xs font-medium flex items-center gap-2">
          <AlertCircle className="w-5 h-5 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Walkthrough Cards */}
      <div className="space-y-4">
        {/* Step 1 */}
        <div className="p-5 rounded-2xl bg-card border border-card-border space-y-2">
          <div className="flex items-center justify-between">
            <h4 className="text-sm font-bold text-white flex items-center gap-2">
              <span className="w-6 h-6 rounded-full bg-slate-800 text-slate-300 flex items-center justify-center text-xs font-mono">
                1
              </span>
              Create an Application in Discord Developer Portal
            </h4>
            <a
              href="https://discord.com/developers/applications"
              target="_blank"
              rel="noopener noreferrer"
              className="text-xs text-blurple hover:underline flex items-center gap-1 font-medium"
            >
              Open Developer Portal <ExternalLink className="w-3 h-3" />
            </a>
          </div>
          <p className="text-xs text-slate-400 pl-8">
            Click <strong>New Application</strong>, choose a name for your bot, and navigate to the <strong>Bot</strong> tab on the left.
          </p>
        </div>

        {/* Step 2 */}
        <div className="p-5 rounded-2xl bg-card border border-card-border space-y-2">
          <h4 className="text-sm font-bold text-white flex items-center gap-2">
            <span className="w-6 h-6 rounded-full bg-slate-800 text-slate-300 flex items-center justify-center text-xs font-mono">
              2
            </span>
            Enable Privileged Gateway Intents
          </h4>
          <p className="text-xs text-slate-400 pl-8">
            Under <strong>Privileged Gateway Intents</strong>, toggle ON <strong>Server Members Intent</strong> and{' '}
            <strong>Message Content Intent</strong>. Click <strong>Save Changes</strong>.
          </p>
        </div>

        {/* Step 3 */}
        <div className="p-5 rounded-2xl bg-card border border-card-border space-y-4">
          <h4 className="text-sm font-bold text-white flex items-center gap-2">
            <span className="w-6 h-6 rounded-full bg-slate-800 text-slate-300 flex items-center justify-center text-xs font-mono">
              3
            </span>
            Reset and Paste Your Bot Token
          </h4>

          <form onSubmit={handleSaveToken} className="space-y-4 pl-8">
            <div>
              <label className="text-xs text-slate-400 block mb-1.5 font-medium">Discord Bot Token</label>
              <input
                type="password"
                value={botToken}
                onChange={(e) => setBotToken(e.target.value)}
                placeholder="MTMxMjM0... (Your Discord Bot Token)"
                className="w-full bg-slate-900 border border-slate-700 rounded-xl px-4 py-3 text-xs text-white placeholder-slate-500 font-mono tracking-wider focus:outline-none focus:border-blurple"
              />
              <span className="text-[11px] text-slate-500 mt-1 block">
                Never share this token with anyone. It is encrypted in HashiCorp Vault.
              </span>
            </div>

            <button
              type="submit"
              disabled={loading || !botToken.trim()}
              className="py-3 px-6 rounded-xl bg-blurple hover:bg-blurple-hover disabled:bg-slate-800 disabled:text-slate-500 text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-blurple/25 transition-all"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" /> Provisioning K8s Pod...
                </>
              ) : (
                <>
                  Save Token & Launch Bot <ArrowRight className="w-4 h-4" />
                </>
              )}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
