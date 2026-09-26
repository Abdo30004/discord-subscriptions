'use client';

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Bot,
  ShoppingBag,
  LayoutDashboard,
  ShieldCheck,
  Sparkles,
  LogOut,
  LogIn,
  Server,
  Code,
  Loader2,
} from 'lucide-react';
import { useAuth } from '@/contexts/AuthContext';

export default function Navbar() {
  const pathname = usePathname();
  const {
    user,
    guilds,
    selectedGuild,
    selectGuild,
    isAdmin,
    isLoading,
    loginWithDiscord,
    loginAsDev,
    logout,
  } = useAuth();

  const navLinks = [
    { href: '/', label: 'Overview', icon: Sparkles },
    { href: '/store', label: 'Bot Catalog', icon: ShoppingBag },
    ...(user ? [{ href: '/dashboard', label: 'My Dashboard', icon: LayoutDashboard }] : []),
    ...(user && isAdmin ? [{ href: '/admin', label: 'Admin Hub', icon: ShieldCheck }] : []),
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
          {isLoading ? (
            <div className="flex items-center gap-2 text-xs text-slate-400 font-mono">
              <Loader2 className="w-4 h-4 animate-spin text-blurple" /> Loading...
            </div>
          ) : user ? (
            <div className="flex items-center gap-3">
              {/* Guild Switcher */}
              {guilds.length > 0 && selectedGuild && (
                <div className="hidden sm:flex items-center gap-2 bg-slate-900 border border-slate-700/80 rounded-lg px-2.5 py-1 text-xs text-slate-300">
                  <Server className="w-3.5 h-3.5 text-blurple shrink-0" />
                  <select
                    value={selectedGuild.id}
                    onChange={(e) => {
                      const found = guilds.find((g) => g.id === e.target.value);
                      if (found) selectGuild(found);
                    }}
                    className="bg-transparent text-white font-medium text-xs focus:outline-none cursor-pointer"
                  >
                    {guilds.map((g) => (
                      <option key={g.id} value={g.id} className="bg-slate-900 text-white">
                        {g.name}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              {/* User Avatar & Name */}
              <div className="flex items-center gap-2 pl-2 border-l border-slate-800">
                <div className="w-8 h-8 rounded-full bg-slate-800 border border-slate-700 overflow-hidden flex items-center justify-center text-xs font-bold text-white uppercase">
                  {user.avatar ? (
                    <img
                      src={`https://cdn.discordapp.com/avatars/${user.id}/${user.avatar}.png`}
                      alt={user.username}
                      className="w-full h-full object-cover"
                      onError={(e) => {
                        (e.target as HTMLElement).style.display = 'none';
                      }}
                    />
                  ) : (
                    <span>{user.username.slice(0, 2)}</span>
                  )}
                </div>
                <div className="hidden lg:block text-left">
                  <p className="text-xs font-semibold text-white leading-tight">
                    {user.global_name || user.username}
                  </p>
                  <p className="text-[10px] text-slate-400 font-mono">@{user.username}</p>
                </div>

                <button
                  type="button"
                  onClick={logout}
                  title="Sign out"
                  className="p-1.5 rounded-lg text-slate-400 hover:text-red-400 hover:bg-red-500/10 transition-colors ml-1"
                >
                  <LogOut className="w-4 h-4" />
                </button>
              </div>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={() => loginAsDev()}
                className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium border border-slate-700 transition-all flex items-center gap-1.5"
                title="Quick login for local testing"
              >
                <Code className="w-3.5 h-3.5 text-emerald-400" /> Dev Login
              </button>
              <button
                type="button"
                onClick={() => loginWithDiscord()}
                className="px-3.5 py-1.5 rounded-lg bg-blurple hover:bg-blurple-hover text-white text-xs font-semibold transition-all shadow-sm flex items-center gap-1.5"
              >
                <LogIn className="w-3.5 h-3.5" /> Login with Discord
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
