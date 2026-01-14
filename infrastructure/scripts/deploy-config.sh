#!/bin/bash
# Деплой конфигураций на VPS

set -e

VPS_HOST="${VPS_HOST:-144.31.199.226}"
VPS_USER="${VPS_USER:-root}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIGS_DIR="$SCRIPT_DIR/../configs"

echo "🚀 Deploying configs to $VPS_HOST..."

# Деплой staging config
echo "📝 Deploying staging config..."
scp "$CONFIGS_DIR/staging.config.json" \
  "$VPS_USER@$VPS_HOST:/var/www/tma/staging/config.json"

# Деплой production config
echo "📝 Deploying production config..."
scp "$CONFIGS_DIR/production.config.json" \
  "$VPS_USER@$VPS_HOST:/var/www/tma/production/config.json"

# Установка прав
ssh "$VPS_USER@$VPS_HOST" << 'EOF'
chown www-data:www-data /var/www/tma/staging/config.json
chown www-data:www-data /var/www/tma/production/config.json
chmod 644 /var/www/tma/staging/config.json
chmod 644 /var/www/tma/production/config.json
echo "✅ Configs deployed and permissions set"
EOF

echo "✅ Config deployment complete!"
