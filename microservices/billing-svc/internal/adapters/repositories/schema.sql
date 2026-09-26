-- Subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    guild_id VARCHAR(64) NOT NULL,
    plan_id VARCHAR(64) NOT NULL,
    bot_type VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,
    provider VARCHAR(32) NOT NULL,
    external_sub_id VARCHAR(128),
    is_dedicated BOOLEAN NOT NULL DEFAULT false,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_guild_id ON subscriptions(guild_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);

-- Promo Codes table
CREATE TABLE IF NOT EXISTS promo_codes (
    id VARCHAR(64) PRIMARY KEY,
    code VARCHAR(64) UNIQUE NOT NULL,
    discount_type VARCHAR(32) NOT NULL,
    discount_value BIGINT NOT NULL,
    max_uses INT NOT NULL DEFAULT 0,
    current_uses INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_promo_codes_code ON promo_codes(code);

-- Voucher / Gift Codes table
CREATE TABLE IF NOT EXISTS voucher_codes (
    id VARCHAR(64) PRIMARY KEY,
    code VARCHAR(64) UNIQUE NOT NULL,
    plan_id VARCHAR(64) NOT NULL,
    bot_type VARCHAR(64) NOT NULL,
    duration_days INT NOT NULL DEFAULT 30,
    is_dedicated BOOLEAN NOT NULL DEFAULT false,
    is_redeemed BOOLEAN NOT NULL DEFAULT false,
    redeemed_by_user_id VARCHAR(64),
    redeemed_by_guild_id VARCHAR(64),
    redeemed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_voucher_codes_code ON voucher_codes(code);

-- Seed initial promo codes for testing
INSERT INTO promo_codes (id, code, discount_type, discount_value, max_uses, current_uses, is_active)
VALUES
('promo-summer50', 'SUMMER50', 'percentage', 50, 100, 0, true),
('promo-free100', 'VIPFREE', 'percentage', 100, 10, 0, true),
('promo-save2', 'SAVE2', 'fixed', 200, 0, 0, true)
ON CONFLICT (code) DO NOTHING;

-- Seed initial gift voucher for testing
INSERT INTO voucher_codes (id, code, plan_id, bot_type, duration_days, is_dedicated, is_redeemed)
VALUES
('vouch-pro-music', 'GIFT-MUSIC-PRO-30D', 'plan-music-pro', 'music', 30, true, false)
ON CONFLICT (code) DO NOTHING;
