#!/bin/bash

# Setup script for VPS Nginx reverse proxy with SSL
# Run this on your VPS: bash setup-vps-nginx.sh

set -e

echo "=== Setting up Nginx reverse proxy with SSL ==="

# Update system
echo "Updating system packages..."
sudo apt update

# Install Nginx
echo "Installing Nginx..."
sudo apt install -y nginx

# Install Certbot for Let's Encrypt SSL
echo "Installing Certbot..."
sudo apt install -y certbot python3-certbot-nginx

# Stop Nginx temporarily for initial SSL setup
sudo systemctl stop nginx

# Obtain SSL certificate
echo "Obtaining SSL certificate for dev.quiz-sprint-tma.online..."
sudo certbot certonly --standalone -d dev.quiz-sprint-tma.online --non-interactive --agree-tos --email admin@quiz-sprint-tma.online

# Copy Nginx configuration
echo "Setting up Nginx configuration..."
sudo tee /etc/nginx/sites-available/dev-tma > /dev/null <<'EOF'
server {
    listen 80;
    server_name dev.quiz-sprint-tma.online;

    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name dev.quiz-sprint-tma.online;

    # SSL Certificate paths
    ssl_certificate /etc/letsencrypt/live/dev.quiz-sprint-tma.online/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/dev.quiz-sprint-tma.online/privkey.pem;

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Logs
    access_log /var/log/nginx/dev-tma-access.log;
    error_log /var/log/nginx/dev-tma-error.log;

    # Proxy to local dev server (tunneled from MacBook)
    location / {
        proxy_pass http://127.0.0.1:5173;
        proxy_http_version 1.1;

        # WebSocket support for HMR
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Standard proxy headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts for WebSocket
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }
}
EOF

# Enable the site
sudo ln -sf /etc/nginx/sites-available/dev-tma /etc/nginx/sites-enabled/

# Remove default site if exists
sudo rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
echo "Testing Nginx configuration..."
sudo nginx -t

# Start Nginx
echo "Starting Nginx..."
sudo systemctl start nginx
sudo systemctl enable nginx

# Setup auto-renewal for SSL certificate
echo "Setting up SSL certificate auto-renewal..."
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer

echo ""
echo "=== Setup complete! ==="
echo ""
echo "Next steps:"
echo "1. Make sure DNS for dev.quiz-sprint-tma.online points to 144.31.199.226"
echo "2. Run the SSH tunnel from your MacBook: ./start-dev-tunnel.sh"
echo "3. Start your Vue dev server on MacBook: npm run dev"
echo ""
