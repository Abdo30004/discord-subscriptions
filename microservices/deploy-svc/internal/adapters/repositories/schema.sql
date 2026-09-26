CREATE TABLE IF NOT EXISTS deployments (
    id VARCHAR(64) PRIMARY KEY,
    subscription_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    guild_id VARCHAR(64) NOT NULL,
    bot_type VARCHAR(64) NOT NULL,
    k8s_namespace VARCHAR(64) NOT NULL,
    k8s_deployment_name VARCHAR(128) NOT NULL,
    image_name VARCHAR(255) NOT NULL,
    image_tag VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_deployments_guild_id ON deployments(guild_id);
CREATE INDEX IF NOT EXISTS idx_deployments_subscription_id ON deployments(subscription_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
