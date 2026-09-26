'use client';

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Bot, ShoppingBag, LayoutDashboard, ShieldCheck, Sparkles, Terminal } from 'lucide-react';
import { MOCK_USER } from '@/lib/api';

export default function Navbar() {
  const pathname = usePathname();

  const navLinks = [
    { href: '/', label: 'Overview', icon: Sparkles },
    { href: '/store', label: 'Bot Catalog', icon: ShoppingBag },
    { href: '/dashboard', label: 'My Dashboard', icon: LayoutDashboard },
    { href: '/admin', label: 'Admin Hub', icon: ShieldCheck },
  ];

  return (
    <header className="sticky top-0 z-50 w-full border-b border-card-border bg-background/80 backdrop-blur-xl">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Brand Logo */}
        <Link href="/" className="flex items-center gap-3 group">
          <div className="w-10 h-10 rounded-xl bg-blurple flex items-center justify-center text-white shadow-lg shadow-blurple/30 group-hover:scale-105 transition-transform">
            <Bot className="w-6 h-6" />
          </div>
          <div>
            <span className="font-bold text-lg text-white tracking-tight flex items-center gap-1.5">
              Discord<span className="text-blurple">Bots</span>
              <span className="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-blurple/20 text-blurple border border-blurple/30">
                Cloud
              </span>
            </span>
            <p className="text-[11px] text-slate-400 font-mono">Microservices Orchestrator</p>
          </div>
        </Link>

        {/* Navigation Links */}
        <nav className="hidden md:flex items-center gap-1">
          {navLinks.map((link) => {
            const Icon = link.icon;
            const isActive = pathname === link.href;
            return (
              <Link
                key={link.href}
                href={link.href}
                className={`flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm font-medium transition-all ${
                  isActive
                    ? 'bg-blurple/15 text-white border border-blurple/30 shadow-sm'
                    : 'text-slate-400 hover:text-white hover:bg-slate-800/50'
                }`}
              >
                <Icon className={`w-4 h-4 ${isActive ? 'text-blurple' : 'text-slate-400'}`} />
                {link.label}
              </Link>
            );
          })}
        </nav>

        {/* User Account / Discord Status */}
        <div className="flex items-center gap-3">
          <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-xs font-medium text-emerald-400">
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
            K8s Cluster Active
          </div>

          <div className="flex items-center gap-3 pl-3 border-l border-slate-800">
            <div className="w-9 h-9 rounded-full bg-gradient-to-tr from-blurple to-indigo-500 p-0.5">
              <div className="w-full h-full rounded-full bg-slate-900 flex items-center justify-center text-xs font-bold text-white uppercase">
                {MOCK_USER.username.slice(0, 2)}
              </div>
            </div>
            <div className="hidden lg:block text-left">
              <p className="text-xs font-semibold text-white leading-tight">{MOCK_USER.global_name}</p>
              <p className="text-[11px] text-slate-400 font-mono">@{MOCK_USER.username}</p>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}
