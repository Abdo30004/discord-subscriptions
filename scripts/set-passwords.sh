#!/bin/sh
set -e

# Update PostgreSQL per-service user passwords from environment variables if provided
if [ -n "$AUTH_DB_PASS" ]; then
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "ALTER USER auth_user WITH ENCRYPTED PASSWORD '$AUTH_DB_PASS';"
fi
if [ -n "$CATALOG_DB_PASS" ]; then
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "ALTER USER catalog_user WITH ENCRYPTED PASSWORD '$CATALOG_DB_PASS';"
fi
if [ -n "$BILLING_DB_PASS" ]; then
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "ALTER USER billing_user WITH ENCRYPTED PASSWORD '$BILLING_DB_PASS';"
fi
if [ -n "$DEPLOY_DB_PASS" ]; then
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "ALTER USER deploy_user WITH ENCRYPTED PASSWORD '$DEPLOY_DB_PASS';"
fi
if [ -n "$MONITOR_DB_PASS" ]; then
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "ALTER USER monitor_user WITH ENCRYPTED PASSWORD '$MONITOR_DB_PASS';"
fi
