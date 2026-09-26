-- Bot Templates table
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

-- Subscription Plans table
CREATE TABLE IF NOT EXISTS subscription_plans (
    id VARCHAR(64) PRIMARY KEY,
    bot_id VARCHAR(64) NOT NULL REFERENCES bot_templates(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    description TEXT NOT NULL,
    interval VARCHAR(16) NOT NULL, -- 'monthly', 'yearly'
    price_cents BIGINT NOT NULL,
    currency VARCHAR(8) NOT NULL DEFAULT 'USD',
    is_dedicated BOOLEAN NOT NULL DEFAULT false,
    features JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_subscription_plans_bot_id ON subscription_plans(bot_id);

-- Seed Initial Bot Templates
INSERT INTO bot_templates (id, slug, name, description, category, docker_image, default_image_tag, supports_dedicated, supports_shared)
VALUES 
('bot-music-01', 'groovestream', 'GrooveStream Music', 'High-fidelity Discord music bot with Spotify, YouTube, and SoundCloud playback with DSP filters.', 'music', 'ghcr.io/discord-subscriptions/music-bot', 'v1.2.0', true, true),
('bot-mod-02', 'aegis-guardian', 'Aegis Guardian', 'Next-gen moderation bot with AI-powered anti-raid, verification gates, and audit logging.', 'moderation', 'ghcr.io/discord-subscriptions/mod-bot', 'v2.0.1', true, true),
('bot-game-03', 'dungeonquest', 'DungeonQuest RPG', 'Turn-based multiplayer role-playing game bot with guilds, boss raids, and item trading.', 'game', 'ghcr.io/discord-subscriptions/rpg-bot', 'v1.0.0', false, true)
ON CONFLICT (id) DO NOTHING;

-- Seed Initial Plans
INSERT INTO subscription_plans (id, bot_id, name, description, interval, price_cents, is_dedicated, features)
VALUES
('plan-music-free', 'bot-music-01', 'Shared Community', 'Shared cluster music bot with 128kbps audio quality.', 'monthly', 0, false, '["128kbps Audio", "Standard Queue", "Shared Instance"]'::jsonb),
('plan-music-pro', 'bot-music-01', 'Pro Dedicated', 'Dedicated single-tenant instance with 320kbps, 24/7 mode, and custom bot token/avatar.', 'monthly', 799, true, '["320kbps Ultra-HD", "24/7 Always Connected", "Dedicated K8s Pod", "Custom Bot Avatar & Name", "Bass Boost & Equalizer"]'::jsonb),
('plan-mod-pro', 'bot-mod-02', 'Dedicated Defense', 'Dedicated moderation container with sub-millisecond automated raid mitigations.', 'monthly', 999, true, '["Zero Latency Filter", "Dedicated Memory & CPU", "Unlimited Logging Retention", "Custom Bot Persona"]'::jsonb),
('plan-game-guild', 'bot-game-03', 'Server RPG Pass', 'Unlock multiplayer world bosses and x2 XP drops for the entire guild.', 'monthly', 499, false, '["2x Guild XP Multiplier", "Daily Boss Raids", "Exclusive Guild Cosmetics"]'::jsonb)
ON CONFLICT (id) DO NOTHING;
