'use client';

import React, { Component, ErrorInfo, ReactNode } from 'react';
import { AlertCircle, RotateCcw } from 'lucide-react';

interface Props {
  children: ReactNode;
  fallbackTitle?: string;
  fallbackMessage?: string;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Widget error caught by ErrorBoundary component:', error, errorInfo);
  }

  private handleRetry = () => {
    this.setState({ hasError: false, error: null });
  };

  public render() {
    if (this.state.hasError) {
      return (
        <div className="bg-slate-900/90 border border-red-500/30 rounded-2xl p-6 text-center my-4">
          <div className="w-10 h-10 bg-red-500/10 text-red-400 rounded-xl flex items-center justify-center mx-auto mb-3 border border-red-500/20">
            <AlertCircle className="w-5 h-5" />
          </div>
          <h3 className="text-base font-semibold text-white mb-1">
            {this.props.fallbackTitle || 'Component Error'}
          </h3>
          <p className="text-xs text-slate-400 mb-4 max-w-sm mx-auto">
            {this.props.fallbackMessage ||
              this.state.error?.message ||
              'This module encountered an issue while rendering.'}
          </p>
          <button
            onClick={this.handleRetry}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium transition-colors border border-slate-700"
          >
            <RotateCcw className="w-3.5 h-3.5" /> Retry
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
