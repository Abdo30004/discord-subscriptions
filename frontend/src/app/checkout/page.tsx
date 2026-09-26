'use client';

import React, { useState, Suspense } from 'react';
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
} from 'lucide-react';
import { MOCK_BOTS, MOCK_GUILDS, MOCK_USER, initiateCheckout, redeemVoucherCode } from '@/lib/api';

export default function CheckoutPage() {
  return (
    <Suspense fallback={<div className="max-w-4xl mx-auto py-12 px-4 text-center text-slate-400 text-sm">Loading Checkout...</div>}>
      <CheckoutContent />
    </Suspense>
  );
}

function CheckoutContent() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const botId = searchParams.get('bot') || MOCK_BOTS[0].id;
  const botType = searchParams.get('botType') || 'music';
  const planId = searchParams.get('plan') || 'plan-music-pro';
  const guildId = searchParams.get('guild') || MOCK_GUILDS[0].id;
  const isZeroSetup = searchParams.get('zeroSetup') === 'true';

  const bot = MOCK_BOTS.find((b) => b.id === botId) || MOCK_BOTS[0];
  const guild = MOCK_GUILDS.find((g) => g.id === guildId) || MOCK_GUILDS[0];

  const basePriceCents = planId.includes('free') ? 0 : 799;
  const zeroSetupFeeCents = isZeroSetup ? 299 : 0;
  const subtotalCents = basePriceCents + zeroSetupFeeCents;

  const [activeTab, setActiveTab] = useState<'paypal' | 'voucher'>('paypal');
  const [instanceLabel, setInstanceLabel] = useState('');
  const [promoCode, setPromoCode] = useState('');
  const [voucherCode, setVoucherCode] = useState('');
  const [appliedPromo, setAppliedPromo] = useState<string | null>(null);
  const [discountCents, setDiscountCents] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const finalPriceCents = Math.max(0, subtotalCents - discountCents);

  const handleApplyPromo = () => {
    setError(null);
    const upper = promoCode.trim().toUpperCase();
    if (upper === 'VIPFREE') {
      setAppliedPromo('VIPFREE');
      setDiscountCents(subtotalCents); // 100% off
    } else if (upper === 'SUMMER50') {
      setAppliedPromo('SUMMER50');
      setDiscountCents(Math.round(subtotalCents * 0.5)); // 50% off
    } else if (upper === 'SAVE2') {
      setAppliedPromo('SAVE2');
      setDiscountCents(200); // $2 off
    } else {
      setError('Invalid or expired promo code');
    }
  };

  const handleCheckoutSubmit = async () => {
    setLoading(true);
    setError(null);
    try {
      const resp = await initiateCheckout({
        user_id: MOCK_USER.id,
        guild_id: guild.id,
        plan_id: planId,
        bot_type: botType,
        instance_label: instanceLabel.trim() || undefined,
        promo_code: appliedPromo || undefined,
        is_zero_setup: isZeroSetup,
      });

      if (resp.is_free_instant_active && resp.subscription) {
        if (isZeroSetup) {
          router.push(`/dashboard?activated=${resp.subscription.id}&zeroSetup=true`);
        } else {
          router.push(`/setup/${resp.subscription.id}`);
        }
      } else if (resp.approval_url) {
        window.location.href = resp.approval_url;
      }
    } catch (err: any) {
      setError(err.message || 'Payment initiation failed');
    } finally {
      setLoading(false);
    }
  };

  const handleRedeemVoucher = async () => {
    if (!voucherCode.trim()) {
      setError('Please enter a voucher code');
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const sub = await redeemVoucherCode(voucherCode.trim(), MOCK_USER.id, guild.id);
      router.push(`/dashboard?voucherRedeemed=${sub.id}`);
    } catch (err: any) {
      // In dev simulation, if network fails, proceed with mock subscription
      router.push(`/dashboard?voucherRedeemed=sub-gift-test`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto py-12 px-4 sm:px-6 lg:px-8 space-y-8">
      {/* Title */}
      <div>
        <h1 className="text-3xl font-extrabold text-white tracking-tight">Review & Checkout</h1>
        <p className="text-slate-400 text-sm mt-1">
          Complete your bot subscription to begin Kubernetes provisioning for{' '}
          <span className="text-white font-semibold">{guild.name}</span>.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        {/* Left Column: Payment & Code Entry */}
        <div className="md:col-span-2 space-y-6">
          {/* Method Tabs */}
          <div className="flex rounded-xl bg-card border border-card-border p-1">
            <button
              type="button"
              onClick={() => setActiveTab('paypal')}
              className={`flex-1 py-2.5 rounded-lg text-xs font-semibold flex items-center justify-center gap-2 transition-all ${
                activeTab === 'paypal' ? 'bg-blurple text-white shadow-sm' : 'text-slate-400 hover:text-white'
              }`}
            >
              <CreditCard className="w-4 h-4" /> PayPal / Card Subscription
            </button>
            <button
              type="button"
              onClick={() => setActiveTab('voucher')}
              className={`flex-1 py-2.5 rounded-lg text-xs font-semibold flex items-center justify-center gap-2 transition-all ${
                activeTab === 'voucher' ? 'bg-blurple text-white shadow-sm' : 'text-slate-400 hover:text-white'
              }`}
            >
              <Gift className="w-4 h-4" /> Redeem Gift Voucher
            </button>
          </div>

          {error && (
            <div className="p-3.5 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-center gap-2">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {activeTab === 'paypal' ? (
            <div className="space-y-6">
              {/* Instance Label Input */}
              <div className="p-5 rounded-2xl bg-card border border-card-border space-y-2">
                <label className="text-xs font-bold uppercase tracking-wider text-slate-400">
                  Bot Instance Label (Optional)
                </label>
                <p className="text-xs text-slate-500">
                  Assign a unique label if your server runs multiple bots (e.g. "Main Stage", "VIP Lounge", "Anti-Raid Alpha").
                </p>
                <input
                  type="text"
                  value={instanceLabel}
                  onChange={(e) => setInstanceLabel(e.target.value)}
                  placeholder="e.g. Main Stage, VIP Lounge (Auto-assigned if empty)"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3.5 py-2.5 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blurple"
                />
              </div>

              {/* Promo Code Input */}
              <div className="p-5 rounded-2xl bg-card border border-card-border space-y-3">
                <label className="text-xs font-bold uppercase tracking-wider text-slate-400">
                  Have a Promo or Discount Code?
                </label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={promoCode}
                    onChange={(e) => setPromoCode(e.target.value)}
                    placeholder="e.g. VIPFREE, SUMMER50, SAVE2"
                    className="flex-1 bg-slate-900 border border-slate-700 rounded-lg px-3.5 py-2.5 text-xs text-white placeholder-slate-500 uppercase tracking-wider font-mono focus:outline-none focus:border-blurple"
                  />
                  <button
                    type="button"
                    onClick={handleApplyPromo}
                    className="px-4 py-2.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-white text-xs font-semibold border border-slate-700 transition-colors"
                  >
                    Apply
                  </button>
                </div>

                {appliedPromo && (
                  <div className="flex items-center justify-between text-xs text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-3 py-2 rounded-lg">
                    <span className="flex items-center gap-1.5 font-semibold">
                      <CheckCircle2 className="w-4 h-4" /> Code {appliedPromo} applied
                    </span>
                    <span className="font-mono font-bold">-${(discountCents / 100).toFixed(2)}</span>
                  </div>
                )}
              </div>

              {/* Action Button */}
              <button
                type="button"
                onClick={handleCheckoutSubmit}
                disabled={loading}
                className={`w-full py-4 px-6 rounded-xl font-bold text-sm flex items-center justify-center gap-2 shadow-xl transition-all ${
                  finalPriceCents === 0
                    ? 'bg-emerald-500 hover:bg-emerald-600 text-slate-950 shadow-emerald-500/25'
                    : 'bg-[#0070ba] hover:bg-[#005ea6] text-white shadow-[#0070ba]/25'
                }`}
              >
                {loading ? (
                  <>
                    <Loader2 className="w-5 h-5 animate-spin" /> Provisioning Cluster...
                  </>
                ) : finalPriceCents === 0 ? (
                  <>
                    <Zap className="w-5 h-5" /> Activate 100% Free Subscription ($0.00)
                  </>
                ) : (
                  <>
                    <Lock className="w-4 h-4" /> Subscribe with PayPal (${(finalPriceCents / 100).toFixed(2)}/mo)
                  </>
                )}
              </button>

              <div className="flex items-center justify-center gap-2 text-xs text-slate-500">
                <ShieldCheck className="w-4 h-4 text-emerald-400" />
                <span>256-bit encrypted checkout. Automatic monthly renewal. Cancel anytime.</span>
              </div>
            </div>
          ) : (
            /* Voucher Tab */
            <div className="p-6 rounded-2xl bg-card border border-card-border space-y-4">
              <div className="space-y-1">
                <h4 className="text-sm font-bold text-white">Enter Your Gift Card / Voucher</h4>
                <p className="text-xs text-slate-400">
                  Vouchers grant prepaid access to bot features without requiring a credit card or PayPal account.
                </p>
              </div>

              <div className="space-y-3">
                <input
                  type="text"
                  value={voucherCode}
                  onChange={(e) => setVoucherCode(e.target.value)}
                  placeholder="e.g. GIFT-MUSIC-PRO-30D"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-3 text-sm text-white placeholder-slate-500 font-mono tracking-wider uppercase focus:outline-none focus:border-blurple"
                />

                <button
                  type="button"
                  onClick={handleRedeemVoucher}
                  disabled={loading}
                  className="w-full py-3.5 px-4 rounded-xl bg-blurple hover:bg-blurple-hover text-white text-xs font-bold flex items-center justify-center gap-2 shadow-lg shadow-blurple/25 transition-all"
                >
                  {loading ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <>
                      Redeem & Launch Bot <ArrowRight className="w-4 h-4" />
                    </>
                  )}
                </button>
              </div>
            </div>
          )}
        </div>

        {/* Right Column: Order Summary Card */}
        <div className="space-y-4">
          <div className="glass-card rounded-2xl p-6 space-y-5 border border-card-border">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">Order Summary</h3>

            <div className="space-y-3">
              <div>
                <h4 className="text-sm font-bold text-white">{bot.name}</h4>
                <p className="text-xs text-slate-400">Target: {guild.name}</p>
              </div>

              {isZeroSetup && (
                <div className="flex items-center gap-2 px-2.5 py-1.5 rounded-lg bg-amber-500/10 border border-amber-500/20 text-[11px] text-amber-300 font-medium">
                  <Zap className="w-3.5 h-3.5 shrink-0" />
                  <span>Turnkey 0-Setup Enabled</span>
                </div>
              )}
            </div>

            <div className="pt-4 border-t border-slate-800 space-y-2 text-xs">
              <div className="flex justify-between text-slate-400">
                <span>Base Subscription</span>
                <span className="text-white font-mono">${(basePriceCents / 100).toFixed(2)}</span>
              </div>

              {isZeroSetup && (
                <div className="flex justify-between text-slate-400">
                  <span>Turnkey Managed Token</span>
                  <span className="text-white font-mono">${(zeroSetupFeeCents / 100).toFixed(2)}</span>
                </div>
              )}

              {discountCents > 0 && (
                <div className="flex justify-between text-emerald-400 font-medium">
                  <span>Promo Discount</span>
                  <span className="font-mono">-${(discountCents / 100).toFixed(2)}</span>
                </div>
              )}

              <div className="flex justify-between text-base font-bold text-white pt-3 border-t border-slate-800">
                <span>Total Due</span>
                <span className="text-blurple font-mono text-xl">${(finalPriceCents / 100).toFixed(2)}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
