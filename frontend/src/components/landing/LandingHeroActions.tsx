'use client';

import React, { Suspense } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import { ArrowRight, LogIn, LayoutDashboard, Lock, Shield } from 'lucide-react';

export function LandingHeroActions() {
  return (
    <Suspense fallback={<div className="h-14" />}>
      <LandingHeroActionsContent />
    </Suspense>
  );
}

function LandingHeroActionsContent() {
  const searchParams = useSearchParams();
  const authParam = searchParams.get('auth');
  const { user, loginWithDiscord } = useAuth();

  return (
    <>
      {/* Auth Alerts when redirected */}
      {authParam === 'required' && (
        <div className="max-w-3xl mx-auto p-4 rounded-xl bg-blurple/15 border border-blurple/40 text-left flex flex-col sm:flex-row sm:items-center justify-between gap-4 shadow-lg shadow-blurple/10 mb-8">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-lg bg-blurple/20 text-blurple shrink-0">
              <Lock className="w-5 h-5" />
            </div>
            <div>
              <h4 className="text-sm font-bold text-white">Sign In Required</h4>
              <p className="text-xs text-slate-300">
                You need to sign in with Discord to access the Bot Fleet Dashboard and manage your servers.
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            <button
              type="button"
              onClick={() => loginWithDiscord()}
              className="px-4 py-2 rounded-lg bg-blurple hover:bg-blurple-hover text-white text-xs font-bold transition-all shadow flex items-center gap-2"
            >
              <LogIn className="w-4 h-4" /> Sign In with Discord
            </button>
          </div>
        </div>
      )}

      {authParam === 'admin_required' && (
        <div className="max-w-3xl mx-auto p-4 rounded-xl bg-rose-500/15 border border-rose-500/40 text-left flex items-center gap-3 shadow-lg shadow-rose-500/10 mb-8">
          <div className="p-2.5 rounded-lg bg-rose-500/20 text-rose-400 shrink-0">
            <Shield className="w-5 h-5" />
          </div>
          <div>
            <h4 className="text-sm font-bold text-white">Administrator Access Restricted</h4>
            <p className="text-xs text-slate-300">
              The Admin Command Hub is restricted to verified platform administrators.
            </p>
          </div>
        </div>
      )}

      {/* Hero Action Buttons */}
      <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
        <Link
          href="/store"
          className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-8 py-3.5 rounded-xl bg-gradient-to-r from-blurple to-indigo-600 hover:from-blurple-hover hover:to-indigo-500 text-white font-semibold text-base shadow-lg shadow-blurple/25 transition-all transform hover:-translate-y-0.5"
        >
          Explore Bot Store <ArrowRight className="w-4 h-4" />
        </Link>

        {user ? (
          <Link
            href="/dashboard"
            className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-8 py-3.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 font-semibold text-base border border-slate-700 hover:border-slate-600 transition-all"
          >
            <LayoutDashboard className="w-4 h-4 text-blurple" /> Open Fleet Dashboard
          </Link>
        ) : (
          <button
            type="button"
            onClick={() => loginWithDiscord()}
            className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-8 py-3.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 font-semibold text-base border border-slate-700 hover:border-slate-600 transition-all"
          >
            <LogIn className="w-4 h-4 text-blurple" /> Sign In with Discord
          </button>
        )}
      </div>
    </>
  );
}

export default LandingHeroActions;
