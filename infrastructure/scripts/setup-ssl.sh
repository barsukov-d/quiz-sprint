#!/bin/bash
set -e

echo "🔐 Setting up SSL certificates..."

# Проверка root
if [[ $EUID -ne 0 ]]; then
   echo "❌ This script must be run as root"
   exit 1
fi

EMAIL="barsukov.d@gmail.com"

# Проверка DNS перед получением сертификата
echo "🔍 Checking DNS records..."

check_dns() {
    local domain=$1
    echo "Checking $domain..."

    if host $domain > /dev/null 2>&1; then
        RESOLVED_IP=$(host $domain | grep "has address" | awk '{print $4}' | head -1)
        SERVER_IP=$(curl -s ifconfig.me)

        if [ "$RESOLVED_IP" == "$SERVER_IP" ]; then
            echo "✅ $domain → $RESOLVED_IP (correct)"
            return 0
        else
            echo "⚠️  $domain → $RESOLVED_IP (expected: $SERVER_IP)"
            return 1
        fi
    else
        echo "❌ $domain - DNS not resolved"
        return 1
    fi
}

# Проверка staging domain
STAGING_DNS_OK=false
if check_dns "staging.quiz-sprint-tma.online"; then
    STAGING_DNS_OK=true
fi

# Проверка production domain
PRODUCTION_DNS_OK=false
if check_dns "quiz-sprint-tma.online"; then
    PRODUCTION_DNS_OK=true
fi

echo ""

# Получение SSL для staging
if [ "$STAGING_DNS_OK" = true ]; then
    echo "📝 Obtaining SSL for staging.quiz-sprint-tma.online..."
    certbot --nginx \
        -d staging.quiz-sprint-tma.online \
        --non-interactive \
        --agree-tos \
        -m $EMAIL \
        --redirect
    echo "✅ Staging SSL configured"
else
    echo "⏭️  Skipping staging SSL (DNS not configured)"
fi

echo ""

# Получение SSL для production (опционально)
read -p "Do you want to set up production SSL? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ "$PRODUCTION_DNS_OK" = true ]; then
        echo "📝 Obtaining SSL for quiz-sprint-tma.online..."
        certbot --nginx \
            -d quiz-sprint-tma.online \
            --non-interactive \
            --agree-tos \
            -m $EMAIL \
            --redirect
        echo "✅ Production SSL configured"
    else
        echo "❌ Cannot obtain production SSL - DNS not configured"
    fi
fi

echo ""

# Auto-renewal setup
echo "⏰ Setting up auto-renewal..."
systemctl enable certbot.timer
systemctl start certbot.timer

echo ""
echo "✅ SSL setup complete!"
echo ""
echo "📋 Installed certificates:"
certbot certificates

echo ""
echo "🔄 Auto-renewal status:"
systemctl status certbot.timer --no-pager
