import { config } from '../config';

export interface Deployment {
  id: string;
  subscription_id: string;
  guild_id: string;
  bot_type: string;
  instance_label?: string;
  status: string;
  image_name: string;
  image_tag: string;
  is_zero_setup?: boolean;
  client_id?: string;
  created_at: string;
}

export async function getGuildDeployments(guildId: string): Promise<Deployment[]> {
  try {
    const res = await fetch(`${config.deploySvcUrl}/api/v1/deployments/guild/${guildId}`);
    if (!res.ok) return [];
    const body = await res.json();
    if (Array.isArray(body.deployments)) {
      return body.deployments;
    }
    if (body.data) {
      return Array.isArray(body.data) ? body.data : [body.data];
    }
    return [];
  } catch {
    return [];
  }
}

export async function getGuildDeployment(guildId: string): Promise<Deployment | null> {
  const deps = await getGuildDeployments(guildId);
  return deps.length > 0 ? deps[0] : null;
}

export async function restartDeployment(deploymentId: string): Promise<{ success: boolean; message: string }> {
  try {
    const res = await fetch(`${config.deploySvcUrl}/api/v1/deployments/${deploymentId}/restart`, {
      method: 'POST',
    });
    const body = await res.json();
    if (!res.ok) {
      return { success: false, message: body.error || 'Restart failed' };
    }
    return { success: true, message: body.message || 'Restart initiated' };
  } catch (err: any) {
    return { success: false, message: err.message || 'Connection error' };
  }
}

export async function customizeBot(
  identifier: string,
  options: { name?: string; avatarUrl?: string }
): Promise<{ success: boolean; message: string }> {
  try {
    const url = identifier.startsWith('dep-')
      ? `${config.deploySvcUrl}/api/v1/deployments/${identifier}/customize`
      : `${config.deploySvcUrl}/api/v1/deployments/guild/${identifier}/customize`;

    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: options.name,
        avatar_url: options.avatarUrl,
      }),
    });
    const body = await res.json();
    if (!res.ok) {
      return { success: false, message: body.error || 'Customization failed' };
    }
    return { success: true, message: body.message || 'Appearance updated successfully!' };
  } catch (err: any) {
    return { success: false, message: err.message || 'Network error contacting deploy service' };
  }
}
