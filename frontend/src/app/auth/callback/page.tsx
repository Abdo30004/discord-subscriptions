'use client';

import React, { useEffect, useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Loader2, AlertCircle, CheckCircle2 } from 'lucide-react';
import { authenticateWithDiscord } from '@/lib/api';

export default function AuthCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-[60vh] flex items-center justify-center">
          <div className="flex items-center gap-3 text-slate-300">
            <Loader2 className="w-6 h-6 animate-spin text-blurple" />
            <span>Connecting to Discord...</span>
          </div>
        </div>
      }
    >
      <CallbackContent />
    </Suspense>
  );
}

function CallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<string>('Verifying Discord credentials...');

  useEffect(() => {
    const code = searchParams.get('code');
    const authError = searchParams.get('error_description') || searchParams.get('error');

    if (authError) {
      setError(authError);
      return;
    }

    if (!code) {
      setError('Authorization code missing from callback URL');
      return;
    }

    setStatus('Exchanging code for secure platform session...');
    authenticateWithDiscord(code)
      .then((session) => {
        localStorage.setItem('discord_bot_platform_auth_token', session.token);
        setStatus('Authentication successful! Redirecting to Control Center...');
        setTimeout(() => {
          window.location.href = '/dashboard';
        }, 1000);
      })
      .catch((err) => {
        console.error('OAuth exchange error:', err);
        setError(err.message || 'Discord authentication failed');
      });
  }, [searchParams, router]);

  return (
    <div className="min-h-[60vh] flex items-center justify-center px-4">
      <div className="max-w-md w-full p-8 rounded-2xl bg-card border border-card-border shadow-xl text-center space-y-4">
        {error ? (
          <>
            <div className="w-12 h-12 rounded-full bg-red-500/10 text-red-400 mx-auto flex items-center justify-center">
              <AlertCircle className="w-6 h-6" />
            </div>
            <h2 className="text-lg font-bold text-white">Authentication Failed</h2>
            <p className="text-xs text-slate-400">{error}</p>
            <div className="pt-2">
              <button
                type="button"
                onClick={() => router.push('/')}
                className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-white text-xs font-semibold"
              >
                Return to Home
              </button>
            </div>
          </>
        ) : (
          <>
            <div className="w-12 h-12 rounded-full bg-blurple/10 text-blurple mx-auto flex items-center justify-center">
              <Loader2 className="w-6 h-6 animate-spin" />
            </div>
            <h2 className="text-lg font-bold text-white">Signing In</h2>
            <p className="text-xs text-slate-400">{status}</p>
          </>
        )}
      </div>
    </div>
  );
}
