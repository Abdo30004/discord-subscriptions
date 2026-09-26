-- Initialize separate databases for each microservice to follow database-per-service pattern
CREATE DATABASE auth_db;
CREATE DATABASE billing_db;
CREATE DATABASE catalog_db;
CREATE DATABASE deploy_db;
CREATE DATABASE monitor_db;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE auth_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE billing_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE catalog_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE deploy_db TO postgres;
GRANT ALL PRIVILEGES ON DATABASE monitor_db TO postgres;

-- Setup auth_db schema
\c auth_db;

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(64) PRIMARY KEY,
    username VARCHAR(128) NOT NULL,
    global_name VARCHAR(128),
    avatar VARCHAR(128),
    email VARCHAR(255),
    access_token TEXT,
    refresh_token TEXT,
    token_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Setup catalog_db schema and seeds
\c catalog_db;

CREATE TABLE IF NOT EXISTS bot_templates (
    id VARCHAR(64) PRIMARY KEY,
    slug VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(32) NOT NULL,
    docker_image VARCHAR(255) NOT NULL,
    default_image_tag VARCHAR(64) NOT NULL DEFAULT 'latest',
    supports_dedicated BOOLEAN NOT NULL DEFAULT true,
    supports_shared BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subscription_plans (
    id VARCHAR(64) PRIMARY KEY,
    bot_id VARCHAR(64) NOT NULL REFERENCES bot_templates(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    description TEXT NOT NULL,
    interval VARCHAR(16) NOT NULL,
    price_cents BIGINT NOT NULL,
    currency VARCHAR(8) NOT NULL DEFAULT 'USD',
    is_dedicated BOOLEAN NOT NULL DEFAULT false,
    features JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_subscription_plans_bot_id ON subscription_plans(bot_id);

INSERT INTO bot_templates (id, slug, name, description, category, docker_image, default_image_tag, supports_dedicated, supports_shared)
VALUES 
('bot-music-01', 'groovestream', 'GrooveStream Music', 'High-fidelity Discord music bot with Spotify, YouTube, and SoundCloud playback with DSP filters.', 'music', 'ghcr.io/discord-subscriptions/music-bot', 'v1.2.0', true, true),
('bot-mod-02', 'aegis-guardian', 'Aegis Guardian', 'Next-gen moderation bot with AI-powered anti-raid, verification gates, and audit logging.', 'moderation', 'ghcr.io/discord-subscriptions/mod-bot', 'v2.0.1', true, true),
('bot-game-03', 'dungeonquest', 'DungeonQuest RPG', 'Turn-based multiplayer role-playing game bot with guilds, boss raids, and item trading.', 'game', 'ghcr.io/discord-subscriptions/rpg-bot', 'v1.0.0', false, true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO subscription_plans (id, bot_id, name, description, interval, price_cents, is_dedicated, features)
VALUES
('plan-music-free', 'bot-music-01', 'Shared Community', 'Shared cluster music bot with 128kbps audio quality.', 'monthly', 0, false, '["128kbps Audio", "Standard Queue", "Shared Instance"]'::jsonb),
('plan-music-pro', 'bot-music-01', 'Pro Dedicated', 'Dedicated single-tenant instance with 320kbps, 24/7 mode, and custom bot token/avatar.', 'monthly', 799, true, '["320kbps Ultra-HD", "24/7 Always Connected", "Dedicated K8s Pod", "Custom Bot Avatar & Name", "Bass Boost & Equalizer"]'::jsonb),
('plan-mod-pro', 'bot-mod-02', 'Dedicated Defense', 'Dedicated moderation container with sub-millisecond automated raid mitigations.', 'monthly', 999, true, '["Zero Latency Filter", "Dedicated Memory & CPU", "Unlimited Logging Retention", "Custom Bot Persona"]'::jsonb),
('plan-game-guild', 'bot-game-03', 'Server RPG Pass', 'Unlock multiplayer world bosses and x2 XP drops for the entire guild.', 'monthly', 499, false, '["2x Guild XP Multiplier", "Daily Boss Raids", "Exclusive Guild Cosmetics"]'::jsonb)
ON CONFLICT (id) DO NOTHING;

-- Setup billing_db schema
\c billing_db;

CREATE TABLE IF NOT EXISTS subscriptions (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    guild_id VARCHAR(64) NOT NULL,
    plan_id VARCHAR(64) NOT NULL,
    bot_type VARCHAR(64) NOT NULL,
    instance_label VARCHAR(128) NOT NULL DEFAULT 'Default',
    status VARCHAR(32) NOT NULL,
    provider VARCHAR(32) NOT NULL,
    external_sub_id VARCHAR(128),
    is_dedicated BOOLEAN NOT NULL DEFAULT false,
    is_zero_setup BOOLEAN NOT NULL DEFAULT false,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_guild_id ON subscriptions(guild_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);

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

INSERT INTO promo_codes (id, code, discount_type, discount_value, max_uses, current_uses, is_active)
VALUES 
('promo-summer50', 'SUMMER50', 'percentage', 50, 100, 0, true),
('promo-free100', 'VIPFREE', 'percentage', 100, 10, 0, true),
('promo-save2', 'SAVE2', 'fixed', 200, 0, 0, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO voucher_codes (id, code, plan_id, bot_type, duration_days, is_dedicated, is_redeemed)
VALUES
('vouch-pro-music', 'GIFT-MUSIC-PRO-30D', 'plan-music-pro', 'music', 30, true, false)
ON CONFLICT (code) DO NOTHING;

-- Setup deploy_db schema
\c deploy_db;

CREATE TABLE IF NOT EXISTS deployments (
    id VARCHAR(64) PRIMARY KEY,
    subscription_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    guild_id VARCHAR(64) NOT NULL,
    bot_type VARCHAR(64) NOT NULL,
    instance_label VARCHAR(128) NOT NULL DEFAULT 'Default',
    k8s_namespace VARCHAR(64) NOT NULL,
    k8s_deployment_name VARCHAR(128) NOT NULL,
    image_name VARCHAR(255) NOT NULL,
    image_tag VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,
    is_zero_setup BOOLEAN NOT NULL DEFAULT false,
    client_id VARCHAR(64),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_deployments_guild_id ON deployments(guild_id);
CREATE INDEX IF NOT EXISTS idx_deployments_subscription_id ON deployments(subscription_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);

CREATE TABLE IF NOT EXISTS token_pool (
    id VARCHAR(64) PRIMARY KEY,
    bot_type VARCHAR(64) NOT NULL,
    client_id VARCHAR(64) NOT NULL,
    token_vault_path VARCHAR(255) NOT NULL,
    token_encrypted TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'available',
    assigned_guild_id VARCHAR(64),
    assigned_subscription_id VARCHAR(64),
    assigned_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_token_pool_type_status ON token_pool(bot_type, status);
CREATE INDEX IF NOT EXISTS idx_token_pool_guild_id ON token_pool(assigned_guild_id);

-- Pre-warm managed bot pool with simulated tokens for 0-setup deployment
INSERT INTO token_pool (id, bot_type, client_id, token_vault_path, token_encrypted, status)
VALUES
('pool-music-01', 'music', '131234567890123456', 'secret/data/bots/pool/pool-music-01', 'MTMxMjM0NTY3ODkwMTIzNDU2.Gz9abc.managed_token_music_01_xyz', 'available'),
('pool-music-02', 'music', '131234567890123457', 'secret/data/bots/pool/pool-music-02', 'MTMxMjM0NTY3ODkwMTIzNDU3.Gz9abc.managed_token_music_02_xyz', 'available'),
('pool-mod-01', 'moderation', '131234567890123458', 'secret/data/bots/pool/pool-mod-01', 'MTMxMjM0NTY3ODkwMTIzNDU4.Gz9abc.managed_token_mod_01_xyz', 'available'),
('pool-game-01', 'game', '131234567890123459', 'secret/data/bots/pool/pool-game-01', 'MTMxMjM0NTY3ODkwMTIzNDU5.Gz9abc.managed_token_game_01_xyz', 'available')
ON CONFLICT (id) DO NOTHING;

-- Setup monitor_db schema
\c monitor_db;

CREATE TABLE IF NOT EXISTS monitoring_targets (
    id VARCHAR(64) PRIMARY KEY,
    bot_id VARCHAR(64) UNIQUE NOT NULL,
    guild_id VARCHAR(64) NOT NULL,
    instance_label VARCHAR(128) NOT NULL DEFAULT 'Default',
    health_url VARCHAR(255) NOT NULL,
    poll_interval_sec INT NOT NULL DEFAULT 15,
    is_active BOOLEAN NOT NULL DEFAULT true,
    current_status VARCHAR(32) NOT NULL DEFAULT 'offline',
    consecutive_failures INT NOT NULL DEFAULT 0,
    last_checked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS health_logs (
    id BIGSERIAL PRIMARY KEY,
    target_id VARCHAR(64) NOT NULL REFERENCES monitoring_targets(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL,
    status_code INT NOT NULL,
    latency_ms BIGINT NOT NULL,
    error_message TEXT,
    discord_ping_ms BIGINT DEFAULT 0,
    memory_usage_mb BIGINT DEFAULT 0,
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_monitoring_targets_is_active ON monitoring_targets(is_active);
CREATE INDEX IF NOT EXISTS idx_monitoring_targets_guild_id ON monitoring_targets(guild_id);
CREATE INDEX IF NOT EXISTS idx_health_logs_target_id ON health_logs(target_id, checked_at DESC);
