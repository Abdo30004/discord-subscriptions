import dotenv from 'dotenv';
dotenv.config();

export const config = {
  discordToken: process.env.MANAGER_BOT_TOKEN || '',
  clientId: process.env.MANAGER_BOT_CLIENT_ID || '',
  catalogSvcUrl: process.env.CATALOG_SVC_URL || 'http://localhost:8081',
  billingSvcUrl: process.env.BILLING_SVC_URL || 'http://localhost:8082',
  deploySvcUrl: process.env.DEPLOY_SVC_URL || 'http://localhost:8083',
  monitorSvcUrl: process.env.MONITOR_SVC_URL || 'http://localhost:8084',
  dashboardUrl: process.env.DASHBOARD_URL || 'http://localhost:3000',
};
