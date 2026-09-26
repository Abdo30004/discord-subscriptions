CREATE TABLE IF NOT EXISTS monitoring_targets (
    id VARCHAR(64) PRIMARY KEY,
    bot_id VARCHAR(64) NOT NULL,
    guild_id VARCHAR(64) UNIQUE NOT NULL,
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
