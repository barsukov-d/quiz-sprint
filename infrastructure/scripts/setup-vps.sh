#!/bin/bash
set -e

echo "🚀 Setting up VPS for quiz-sprint TMA..."

# Проверка root
if [[ $EUID -ne 0 ]]; then
   echo "❌ This script must be run as root"
   exit 1
fi

# 1. Обновление системы
echo "📦 Updating system packages..."
apt update && apt upgrade -y

# 2. Установка Nginx
echo "🔧 Installing Nginx..."
apt install -y nginx

# 3. Установка Certbot для SSL
echo "🔐 Installing Certbot..."
apt install -y certbot python3-certbot-nginx

# 4. Создание директорий для статики
echo "📁 Creating directories..."
mkdir -p /var/www/tma/staging
mkdir -p /var/www/tma/production
chown -R www-data:www-data /var/www/tma

# 5. Копирование Nginx конфигураций
echo "⚙️ Copying Nginx configurations..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cp "$SCRIPT_DIR/../nginx/ssl-params.conf" /etc/nginx/snippets/ssl-params.conf
cp "$SCRIPT_DIR/../nginx/sites-available/staging.conf" /etc/nginx/sites-available/staging
cp "$SCRIPT_DIR/../nginx/sites-available/production.conf" /etc/nginx/sites-available/production

# 6. Включение сайтов
echo "🔗 Enabling sites..."
ln -sf /etc/nginx/sites-available/staging /etc/nginx/sites-enabled/
ln -sf /etc/nginx/sites-available/production /etc/nginx/sites-enabled/

# 7. Удаление default сайта
if [ -f /etc/nginx/sites-enabled/default ]; then
    echo "🗑️  Removing default site..."
    rm /etc/nginx/sites-enabled/default
fi

# 8. Тест конфигурации (без SSL сертификатов будет ошибка, но это нормально)
echo "🧪 Testing Nginx configuration..."
nginx -t || echo "⚠️  Nginx test failed (expected before SSL setup)"

# 9. Перезапуск Nginx
echo "🔄 Restarting Nginx..."
systemctl restart nginx
systemctl enable nginx

# 10. Настройка firewall
echo "🔥 Configuring firewall..."
if command -v ufw &> /dev/null; then
    ufw allow 'Nginx Full'
    ufw allow OpenSSH
    echo "✅ Firewall configured"
fi

echo ""
echo "✅ VPS setup complete!"
echo ""
echo "📋 Next steps:"
echo "1. Ensure DNS records point to this server IP:"
echo "   - staging.quiz-sprint-tma.online → $(curl -s ifconfig.me)"
echo "   - quiz-sprint-tma.online → $(curl -s ifconfig.me)"
echo ""
echo "2. Run: ./setup-ssl.sh to obtain SSL certificates"
echo ""
echo "3. Deploy your application files to:"
echo "   - Staging: /var/www/tma/staging"
echo "   - Production: /var/www/tma/production"
