#!/bin/bash
# Деплой обновленной Nginx конфигурации на VPS

set -e

VPS_HOST="${VPS_HOST:-144.31.199.226}"
VPS_USER="${VPS_USER:-root}"

echo "🚀 Deploying Nginx configuration to $VPS_HOST..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NGINX_DIR="$SCRIPT_DIR/../nginx"

# Проверка наличия конфигурационных файлов
if [ ! -f "$NGINX_DIR/sites-available/staging.conf" ]; then
    echo "❌ staging.conf not found"
    exit 1
fi

if [ ! -f "$NGINX_DIR/ssl-params.conf" ]; then
    echo "❌ ssl-params.conf not found"
    exit 1
fi

# Копирование конфигураций
echo "📦 Copying configuration files..."
scp "$NGINX_DIR/sites-available/staging.conf" $VPS_USER@$VPS_HOST:/etc/nginx/sites-available/staging
scp "$NGINX_DIR/sites-available/production.conf" $VPS_USER@$VPS_HOST:/etc/nginx/sites-available/production
scp "$NGINX_DIR/ssl-params.conf" $VPS_USER@$VPS_HOST:/etc/nginx/snippets/ssl-params.conf

echo "🔗 Enabling sites and testing..."

# Тест и перезагрузка на VPS
ssh $VPS_USER@$VPS_HOST << 'EOF'
# Включение сайтов
ln -sf /etc/nginx/sites-available/staging /etc/nginx/sites-enabled/
ln -sf /etc/nginx/sites-available/production /etc/nginx/sites-enabled/

# Тест конфигурации
if nginx -t; then
    echo "✅ Nginx configuration test passed"

    # Перезагрузка Nginx
    systemctl reload nginx
    echo "✅ Nginx reloaded successfully"

    # Показать статус
    systemctl status nginx --no-pager | head -10
else
    echo "❌ Nginx configuration test failed"
    exit 1
fi
EOF

echo ""
echo "✅ Deployment complete!"
echo ""
echo "🌐 Check your sites:"
echo "   - https://staging.quiz-sprint-tma.online"
echo "   - https://quiz-sprint-tma.online"
