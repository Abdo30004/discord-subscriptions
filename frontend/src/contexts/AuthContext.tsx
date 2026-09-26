'use client';

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import {
  UserProfile,
  DiscordGuild,
} from '@/lib/types';
import {
  getDiscordOAuthUrl,
  getCurrentUser,
  getUserGuilds,
} from '@/lib/api';

interface AuthContextType {
  user: UserProfile | null;
  token: string | null;
  guilds: DiscordGuild[];
  selectedGuild: DiscordGuild | null;
  isAdmin: boolean;
  isSuperAdmin: boolean;
  isLoading: boolean;
  error: string | null;
  loginWithDiscord: (redirectUri?: string) => Promise<void>;
  logout: () => void;
  selectGuild: (guild: DiscordGuild) => void;
  refreshSession: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const TOKEN_KEY = 'discord_bot_platform_auth_token';
const SELECTED_GUILD_KEY = 'discord_bot_platform_selected_guild';

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [guilds, setGuilds] = useState<DiscordGuild[]>([]);
  const [selectedGuild, setSelectedGuild] = useState<DiscordGuild | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const setAuthSession = useCallback(async (authToken: string) => {
    try {
      localStorage.setItem(TOKEN_KEY, authToken);
      setToken(authToken);

      // Fetch user profile and manageable guilds concurrently
      const [userProfile, userGuilds] = await Promise.all([
        getCurrentUser(authToken),
        getUserGuilds(authToken),
      ]);

      setUser(userProfile);
      setGuilds(userGuilds);

      // Restore previously selected guild or pick first available
      const savedGuildId = localStorage.getItem(SELECTED_GUILD_KEY);
      const matchedGuild = userGuilds.find((g) => g.id === savedGuildId);
      const active = matchedGuild || userGuilds[0] || null;
      setSelectedGuild(active);
      if (active) {
        localStorage.setItem(SELECTED_GUILD_KEY, active.id);
      }
      setError(null);
    } catch (err: any) {
      console.error('Failed to restore session:', err);
      localStorage.removeItem(TOKEN_KEY);
      setToken(null);
      setUser(null);
      setGuilds([]);
      setSelectedGuild(null);
      setError(err.message || 'Session expired');
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Initialize session on mount
  useEffect(() => {
    const savedToken = localStorage.getItem(TOKEN_KEY);
    if (savedToken) {
      setAuthSession(savedToken);
    } else {
      setIsLoading(false);
    }
  }, [setAuthSession]);

  const loginWithDiscord = async (redirectUri?: string) => {
    try {
      const url = await getDiscordOAuthUrl(redirectUri);
      window.location.href = url;
    } catch (err: any) {
      setError(err.message || 'Failed to initiate Discord OAuth');
      throw err;
    }
  };

  const logout = () => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(SELECTED_GUILD_KEY);
    setToken(null);
    setUser(null);
    setGuilds([]);
    setSelectedGuild(null);
    setError(null);
  };

  const selectGuild = (guild: DiscordGuild) => {
    setSelectedGuild(guild);
    localStorage.setItem(SELECTED_GUILD_KEY, guild.id);
  };

  const refreshSession = async () => {
    if (token) {
      await setAuthSession(token);
    }
  };

  const isSuperAdmin = Boolean(user && user.is_super_admin);
  const isAdmin = Boolean(user && (user.is_admin || user.is_super_admin));

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        guilds,
        selectedGuild,
        isAdmin,
        isSuperAdmin,
        isLoading,
        error,
        loginWithDiscord,
        logout,
        selectGuild,
        refreshSession,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
