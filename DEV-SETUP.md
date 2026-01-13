# Development Setup with Reverse Proxy

This setup allows you to develop your TMA app locally on MacBook and access it via HTTPS with HMR (Hot Module Reload) support.

## Architecture

```
MacBook (localhost:5173) 
    ↓ SSH Tunnel
VPS (127.0.0.1:5173)
    ↓ Nginx Reverse Proxy
Public (https://dev.quiz-sprint-tma.online)
```

## Setup Complete ✓

- ✅ Nginx installed on VPS (144.31.199.226)
- ✅ SSL certificate obtained (expires 2026-04-13)
- ✅ WebSocket support configured for Vue HMR
- ✅ Auto-renewal enabled for SSL certificate

## How to Use

### 1. Start your local Vue dev server

```bash
cd tma
npm run dev
```

Your app should be running on `http://localhost:5173`

### 2. Start the SSH tunnel

In a new terminal from project root:

```bash
./dev-tunnel/start-dev-tunnel.sh
```

This will forward your local port 5173 to the VPS.

### 3. Access your app

Open in browser: **https://dev.quiz-sprint-tma.online**

## Features

- **HTTPS/SSL**: Secure connection with Let's Encrypt certificate
- **WebSocket Support**: Vue HMR works seamlessly
- **Auto-renewal**: SSL certificate renews automatically
- **Keep-alive**: Connection stays alive with automatic reconnection

## Troubleshooting

### Tunnel disconnects

The tunnel has keep-alive settings and will try to reconnect. If it fails:
- Check VPS is accessible: `ssh root@144.31.199.226`
- Restart the tunnel: `./dev-tunnel/start-dev-tunnel.sh`

### HMR not working

1. Make sure your Vue dev server is running
2. Check the tunnel is active
3. Verify Nginx logs: `ssh root@144.31.199.226 'tail -f /var/log/nginx/dev-tma-error.log'`

### SSL certificate issues

Certificate auto-renews via certbot. To manually renew:
```bash
ssh root@144.31.199.226 'sudo certbot renew'
```

## VPS Management

### View Nginx logs
```bash
ssh root@144.31.199.226 'tail -f /var/log/nginx/dev-tma-access.log'
ssh root@144.31.199.226 'tail -f /var/log/nginx/dev-tma-error.log'
```

### Restart Nginx
```bash
ssh root@144.31.199.226 'sudo systemctl restart nginx'
```

### Check Nginx status
```bash
ssh root@144.31.199.226 'sudo systemctl status nginx'
```

### Test tunnel setup
```bash
./dev-tunnel/test-tunnel.sh
```

## DNS Configuration

DNS is already configured:
- Domain: `dev.quiz-sprint-tma.online`
- Type: A Record
- Value: 144.31.199.226

## Files Structure

```
quiz-sprint/
├── DEV-SETUP.md                    # This file
└── dev-tunnel/
    ├── README.md                   # Detailed documentation
    ├── setup-vps-nginx.sh          # VPS setup script (already executed)
    ├── start-dev-tunnel.sh         # SSH tunnel script (use this!)
    └── test-tunnel.sh              # Test script
```

## Quick Start

```bash
# Terminal 1: Start Vue dev server
cd tma && npm run dev

# Terminal 2: Start tunnel
./dev-tunnel/start-dev-tunnel.sh

# Open browser
open https://dev.quiz-sprint-tma.online
```
