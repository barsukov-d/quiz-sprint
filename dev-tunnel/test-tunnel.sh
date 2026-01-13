#!/bin/bash

# Quick test script to verify the tunnel setup

echo "=== Testing Development Tunnel Setup ==="
echo ""

echo "1. Checking DNS resolution..."
DNS_IP=$(nslookup dev.quiz-sprint-tma.online | grep -A1 "Non-authoritative answer" | grep "Address:" | awk '{print $2}')
if [ "$DNS_IP" == "144.31.199.226" ]; then
    echo "   ✓ DNS is correctly pointing to VPS"
else
    echo "   ✗ DNS issue: got $DNS_IP, expected 144.31.199.226"
fi
echo ""

echo "2. Checking VPS accessibility..."
if ssh -o ConnectTimeout=5 root@144.31.199.226 'echo "Connected"' &>/dev/null; then
    echo "   ✓ VPS is accessible via SSH"
else
    echo "   ✗ Cannot connect to VPS"
fi
echo ""

echo "3. Checking Nginx status..."
NGINX_STATUS=$(ssh root@144.31.199.226 'systemctl is-active nginx' 2>/dev/null)
if [ "$NGINX_STATUS" == "active" ]; then
    echo "   ✓ Nginx is running"
else
    echo "   ✗ Nginx is not running"
fi
echo ""

echo "4. Checking SSL certificate..."
CERT_EXPIRY=$(ssh root@144.31.199.226 'sudo certbot certificates 2>/dev/null | grep "Expiry Date"' | head -1)
if [ ! -z "$CERT_EXPIRY" ]; then
    echo "   ✓ SSL certificate is installed"
    echo "     $CERT_EXPIRY"
else
    echo "   ✗ SSL certificate issue"
fi
echo ""

echo "5. Testing HTTPS endpoint..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" https://dev.quiz-sprint-tma.online --max-time 5 2>/dev/null)
if [ "$HTTP_CODE" == "502" ] || [ "$HTTP_CODE" == "503" ]; then
    echo "   ✓ HTTPS is working (502/503 expected without tunnel)"
    echo "     Nginx is waiting for your local dev server"
elif [ "$HTTP_CODE" == "200" ]; then
    echo "   ✓ HTTPS is working and app is responding!"
else
    echo "   ? Got HTTP $HTTP_CODE"
fi
echo ""

echo "=== Next Steps ==="
echo "1. Start your Vue dev server: cd tma && npm run dev"
echo "2. Start the tunnel: ./start-dev-tunnel.sh"
echo "3. Open: https://dev.quiz-sprint-tma.online"
echo ""
