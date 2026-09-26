'use client';

import React from 'react';
import { AlertOctagon, RotateCcw } from 'lucide-react';

interface GlobalErrorProps {
  error: Error & { digest?: string };
  reset: () => void;
}

export default function GlobalError({ error, reset }: GlobalErrorProps) {
  return (
    <html lang="en" className="dark">
      <body className="bg-slate-950 text-slate-100 min-h-screen flex items-center justify-center p-4 antialiased">
        <div className="max-w-md w-full bg-slate-900 border border-red-500/40 rounded-3xl p-8 text-center shadow-2xl">
          <div className="w-16 h-16 bg-red-500/20 text-red-400 rounded-3xl flex items-center justify-center mx-auto mb-6 border border-red-500/30">
            <AlertOctagon className="w-8 h-8" />
          </div>

          <h1 className="text-2xl font-bold text-white mb-2">Critical Application Failure</h1>
          <p className="text-sm text-slate-400 mb-6">
            The platform encountered a root-level error and could not initialize the user interface.
          </p>

          {error?.message && (
            <div className="bg-slate-950 border border-slate-800 rounded-xl p-3 text-xs text-red-300 font-mono mb-6 text-left break-all">
              {error.message}
            </div>
          )}

          <button
            onClick={() => reset()}
            className="w-full inline-flex items-center justify-center gap-2 px-6 py-3 rounded-xl bg-gradient-to-r from-red-600 to-rose-600 hover:from-red-500 hover:to-rose-500 text-white font-medium text-sm transition-all shadow-lg shadow-red-500/20"
          >
            <RotateCcw className="w-4 h-4" /> Reload Platform
          </button>
        </div>
      </body>
    </html>
  );
}
