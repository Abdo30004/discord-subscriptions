'use client';

import React, { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import {
  ShieldCheck,
  Zap,
  CreditCard,
  Gift,
  ArrowRight,
  CheckCircle2,
  AlertCircle,
  Loader2,
  Lock,
  LogIn,
} from 'lucide-react';
import { getBotById, initiateCheckout, redeemVoucherCode } from '@/lib/api';
import { BotTemplate, SubscriptionPlan } from '@/lib/types';
import { useAuth } from '@/contexts/AuthContext';

export default function CheckoutPage() {
  return (
    <Suspense
      fallback={
        <div className="max-w-4xl mx-auto py-12 px-4 text-center text-slate-400 text-sm flex items-center justify-center gap-2">
          <Loader2 className="w-5 h-5 animate-spin text-blurple" /> Loading Checkout...
        </div>
      }
    >
      <CheckoutContent />
    </Suspense>
  );
}

function CheckoutContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { user, token, guilds, loginWithDiscord } = useAuth();

  const botId = searchParams.get('bot') || 'bot-music-01';
  const botType = searchParams.get('botType') || 'music';
  const planId = searchParams.get('plan') || 'plan-music-pro';
  const guildId = searchParams.get('guild') || '';
  const instanceLabel = searchParams.get('instanceLabel') || 'Default';
  const isZeroSetup = searchParams.get('zeroSetup') === 'true';

  const [bot, setBot] = useState<BotTemplate | null>(null);
  const [plan, setPlan] = useState<SubscriptionPlan | null>(null);
  const [loadingBot, setLoadingBot] = useState<boolean>(true);

  const [activeTab, setActiveTab] = useState<'paypal' | 'voucher'>('paypal');
  const [promoCode, setPromoCode] = useState('');
  const [voucherCode, setVoucherCode] = useState('');
  const [appliedPromo, setAppliedPromo] = useState<string | null>(null);
  const [discountCents, setDiscountCents] = useState(0);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  // Fetch real bot details from catalog service
  useEffect(() => {
    getBotById(botId)
      .then((data) => {
        setBot(data);
        const p = data.plans?.find((item) => item.id === planId) || data.plans?.[0] || null;
        setPlan(p);
      })
      .catch((err) => {
        console.error('Failed to load bot for checkout:', err);
        setError(`Failed to retrieve bot template: ${err.message}`);
      })
      .finally(() => {
        setLoadingBot(false);
      });
  }, [botId, planId]);

  const basePriceCents = plan?.price_cents ?? (planId.includes('free') ? 0 : 799);
  const zeroSetupFeeCents = isZeroSetup ? 299 : 0;
  const subtotalCents = basePriceCents + zeroSetupFeeCents;
  const finalPriceCents = Math.max(0, subtotalCents - discountCents);

  // Find guild name if available from user's guilds
  const guildName = guilds.find((g) => g.id === guildId)?.name || `Server ID: ${guildId || 'Not Specified'}`;

  const handleApplyPromo = () => {
    setError(null);
    const upper = promoCode.trim().toUpperCase();
    if (upper === 'VIPFREE') {
      setAppliedPromo('VIPFREE');
      setDiscountCents(subtotalCents);
    } else if (upper === 'SUMMER50') {
      setAppliedPromo('SUMMER50');
      setDiscountCents(Math.round(subtotalCents * 0.5));
    } else if (upper === 'SAVE2') {
      setAppliedPromo('SAVE2');
      setDiscountCents(200);
    } else {
      setError('Promo code not recognized or expired');
    }
  };

  const handleCheckoutSubmit = async () => {
    if (!user) {
      setError('Please log in with Discord before proceeding to checkout.');
      return;
    }
    if (!guildId) {
      setError('Missing target Discord server. Please return to the catalog and choose a server.');
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      const resp = await initiateCheckout(
        {
          user_id: user.id,
          guild_id: guildId,
          plan_id: planId,
          bot_type: botType,
          instance_label: instanceLabel,
          promo_code: appliedPromo || undefined,
          is_zero_setup: isZeroSetup,
        },
        token || undefined
      );

      if (resp.is_free_instant_active && resp.subscription) {
        if (isZeroSetup) {
          router.push(`/dashboard?activated=${resp.subscription.id}&zeroSetup=true`);
        } else {
          router.push(`/setup/${resp.subscription.id}`);
        }
      } else if (resp.approval_url) {
        window.location.href = resp.approval_url;
      } else {
        router.push(`/dashboard?activated=true`);
      }
    } catch (err: any) {
      setError(err.message || 'Payment initiation failed with billing service');
    } finally {
      setSubmitting(false);
    }
  };

  const handleRedeemVoucher = async () => {
    if (!user) {
      setError('Please log in before redeeming a gift code.');
      return;
    }
    if (!voucherCode.trim()) {
      setError('Please enter a voucher code');
      return;
    }
    if (!guildId) {
      setError('Please select a target Discord server before redeeming');
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      const sub = await redeemVoucherCode(voucherCode.trim(), user.id, guildId, token || undefined);
      setSuccessMessage(`Gift voucher redeemed! Subscription activated until ${new Date(sub.valid_until).toLocaleDateString()}`);
      setTimeout(() => {
        router.push(`/dashboard?activated=${sub.id}`);
      }, 1500);
    } catch (err: any) {
      setError(err.message || 'Invalid or already redeemed voucher code');
    } finally {
      setSubmitting(false);
    }
  };

  if (loadingBot) {
    return (
      <div className="max-w-4xl mx-auto py-24 px-4 text-center space-y-4">
        <Loader2 className="w-8 h-8 animate-spin text-blurple mx-auto" />
        <p className="text-slate-400 text-sm">Loading order summary from Catalog & Billing services...</p>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto py-12 px-4 sm:px-6 lg:px-8 space-y-8">
      {/* Header */}
      <div>
        <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-blurple/15 border border-blurple/30 text-blurple text-xs font-semibold uppercase tracking-wider mb-2">
          <ShieldCheck className="w-3.5 h-3.5" /> Secure Checkout
        </div>
        <h1 className="text-3xl font-extrabold text-white tracking-tight">Complete Your Subscription</h1>
        <p className="text-slate-400 text-sm mt-1">
          Review your cluster allocation, apply vouchers, and confirm provisioning.
        </p>
      </div>

      {/* Auth Banner if not signed in */}
      {!user && (
        <div className="p-5 rounded-2xl bg-slate-900 border border-blurple/30 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="space-y-1">
            <h4 className="text-sm font-bold text-white flex items-center gap-2">
              <LogIn className="w-4 h-4 text-blurple" /> Authentication Required
            </h4>
            <p className="text-xs text-slate-400">
              Sign in with Discord to associate this bot subscription with your profile and server.
            </p>
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => loginWithDiscord()}
              className="px-4 py-2 rounded-xl bg-blurple hover:bg-blurple-hover text-white text-xs font-semibold"
            >
              Connect Discord
            </button>
          </div>
        </div>
      )}

      {error && (
        <div className="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-xs font-medium flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {successMessage && (
        <div className="p-4 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-medium flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4 shrink-0" />
          <span>{successMessage}</span>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {/* Left Column: Order Summary */}
        <div className="md:col-span-1 space-y-4">
          <div className="p-6 rounded-2xl bg-card border border-card-border space-y-4">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">Order Summary</h3>

            <div className="space-y-1">
              <h4 className="text-base font-bold text-white">{bot?.name || botId}</h4>
              <p className="text-xs text-slate-400">{plan?.name || planId}</p>
            </div>

            <div className="p-3 rounded-xl bg-slate-900/80 border border-slate-800 space-y-1 text-xs">
              <span className="text-slate-500 block text-[10px] uppercase font-mono">Target Guild:</span>
              <span className="text-slate-200 font-semibold">{guildName}</span>
              <span className="text-slate-500 block text-[10px] uppercase font-mono pt-1">Instance Tag:</span>
              <span className="text-slate-200 font-mono text-[11px]">{instanceLabel}</span>
            </div>

            {isZeroSetup && (
              <div className="flex items-center gap-2 text-xs text-amber-400 bg-amber-500/10 border border-amber-500/20 p-2.5 rounded-xl">
                <Zap className="w-4 h-4 shrink-0" />
                <span>Pre-warmed Turnkey Bot Included</span>
              </div>
            )}

            <div className="pt-4 border-t border-slate-800 space-y-2 text-xs">
              <div className="flex justify-between text-slate-400">
                <span>Base Tier:</span>
                <span className="font-mono text-white">${(basePriceCents / 100).toFixed(2)}</span>
              </div>

              {isZeroSetup && (
                <div className="flex justify-between text-slate-400">
                  <span>0-Setup Fee:</span>
                  <span className="font-mono text-white">$2.99</span>
                </div>
              )}

              {discountCents > 0 && (
                <div className="flex justify-between text-emerald-400">
                  <span>Discount ({appliedPromo}):</span>
                  <span className="font-mono">-${(discountCents / 100).toFixed(2)}</span>
                </div>
              )}

              <div className="pt-2 border-t border-slate-800 flex justify-between text-sm font-bold text-white">
                <span>Total Today:</span>
                <span className="font-mono text-lg text-blurple">${(finalPriceCents / 100).toFixed(2)}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Right Column: Payment & Voucher Redemption Tabs */}
        <div className="md:col-span-2 space-y-6">
          <div className="flex rounded-xl bg-slate-900 p-1 border border-slate-800">
            <button
              type="button"
              onClick={() => setActiveTab('paypal')}
              className={`flex-1 py-2.5 rounded-lg text-xs font-bold transition-all flex items-center justify-center gap-2 ${
                activeTab === 'paypal'
                  ? 'bg-blurple text-white shadow-md'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <CreditCard className="w-4 h-4" /> PayPal / Card Checkout
            </button>
            <button
              type="button"
              onClick={() => setActiveTab('voucher')}
              className={`flex-1 py-2.5 rounded-lg text-xs font-bold transition-all flex items-center justify-center gap-2 ${
                activeTab === 'voucher'
                  ? 'bg-blurple text-white shadow-md'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <Gift className="w-4 h-4" /> Redeem Gift Voucher
            </button>
          </div>

          {activeTab === 'paypal' ? (
            <div className="p-6 rounded-2xl bg-card border border-card-border space-y-6">
              <div className="space-y-3">
                <label className="text-xs font-bold uppercase tracking-wider text-slate-400 block">
                  Promo / Discount Code
                </label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={promoCode}
                    onChange={(e) => setPromoCode(e.target.value)}
                    placeholder="e.g. VIPFREE, SUMMER50, SAVE2"
                    className="flex-1 bg-slate-900 border border-slate-700 rounded-xl px-4 py-2.5 text-xs text-white uppercase font-mono focus:outline-none focus:border-blurple"
                  />
                  <button
                    type="button"
                    onClick={handleApplyPromo}
                    className="px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-xs font-semibold text-white border border-slate-700 transition-all"
                  >
                    Apply
                  </button>
                </div>
              </div>

              <div className="pt-4 border-t border-slate-800 flex items-center justify-between">
                <div className="text-xs text-slate-400 flex items-center gap-1.5">
                  <Lock className="w-3.5 h-3.5 text-emerald-400" />
                  <span>Encrypted 256-bit Checkout</span>
                </div>

                <button
                  type="button"
                  onClick={handleCheckoutSubmit}
                  disabled={submitting || !user}
                  className="py-3 px-8 rounded-xl bg-blurple hover:bg-blurple-hover disabled:bg-slate-800 disabled:text-slate-500 text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-blurple/25 transition-all"
                >
                  {submitting ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" /> Connecting to Billing Service...
                    </>
                  ) : finalPriceCents === 0 ? (
                    <>
                      Activate Free Instance <ArrowRight className="w-4 h-4" />
                    </>
                  ) : (
                    <>
                      Proceed to PayPal (${(finalPriceCents / 100).toFixed(2)}) <ArrowRight className="w-4 h-4" />
                    </>
                  )}
                </button>
              </div>
            </div>
          ) : (
            <div className="p-6 rounded-2xl bg-card border border-card-border space-y-6">
              <div className="space-y-3">
                <label className="text-xs font-bold uppercase tracking-wider text-slate-400 block">
                  Gift Voucher Code
                </label>
                <input
                  type="text"
                  value={voucherCode}
                  onChange={(e) => setVoucherCode(e.target.value)}
                  placeholder="GIFT-XXXX-XXXX"
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-4 py-3 text-xs text-white uppercase font-mono tracking-wider focus:outline-none focus:border-blurple"
                />
                <span className="text-[11px] text-slate-500 block">
                  Enter an admin-issued voucher code (e.g. GIFT-MUSIC-PRO-30D) to activate instantly with zero card details.
                </span>
              </div>

              <div className="pt-4 border-t border-slate-800 flex justify-end">
                <button
                  type="button"
                  onClick={handleRedeemVoucher}
                  disabled={submitting || !voucherCode.trim() || !user}
                  className="py-3 px-8 rounded-xl bg-blurple hover:bg-blurple-hover disabled:bg-slate-800 disabled:text-slate-500 text-white text-xs font-bold flex items-center gap-2 shadow-lg shadow-blurple/25 transition-all"
                >
                  {submitting ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" /> Redeeming Voucher...
                    </>
                  ) : (
                    <>
                      Redeem & Provision Bot <ArrowRight className="w-4 h-4" />
                    </>
                  )}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
