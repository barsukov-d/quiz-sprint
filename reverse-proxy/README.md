# Reverse Proxy for TMA Development

Docker + Caddy reverse proxy with automatic HTTPS for Telegram Mini App development.

## Features

- ✅ Automatic HTTPS with Let's Encrypt
- ✅ Auto-renewal of SSL certificates
- ✅ WebSocket support for Vite HMR (Hot Module Replacement)
- ✅ HTTP/3 support
- ✅ Gzip compression
- ✅ JSON access logs

## Prerequisites

1. **Docker and Docker Compose** installed
2. **Domain name** pointed to your VPS (e.g., `tma.yourdomain.com`)
3. **Ports 80 and 443** open on your VPS firewall

## Setup

### 1. Configure Environment

```bash
cd reverse-proxy
cp .env.example .env
```

Edit `.env`:
```bash
DOMAIN=tma.yourdomain.com  # Replace with your actual domain
VITE_PORT=5173             # Keep default or change if needed
```

### 2. Update Caddyfile Email

Edit `Caddyfile` and replace `your-email@example.com` with your actual email:
```
{
    email your-actual-email@example.com
}
```

### 3. Start Vite Dev Server

Make sure your TMA Vite dev server is running:
```bash
cd ../tma
npm run dev
# Should be running on http://localhost:5173
```

### 4. Start Caddy Reverse Proxy

```bash
cd ../reverse-proxy
docker-compose up -d
```

### 5. Verify

Check logs:
```bash
docker-compose logs -f caddy
```

Your TMA should now be accessible at `https://tma.yourdomain.com`

## Usage

### Start Proxy
```bash
docker-compose up -d
```

### Stop Proxy
```bash
docker-compose down
```

### Restart Proxy
```bash
docker-compose restart
```

### View Logs
```bash
docker-compose logs -f caddy
```

### Access Logs
```bash
docker exec tma-reverse-proxy cat /data/access.log
```

## Testing with Let's Encrypt Staging

To avoid rate limits during testing, uncomment this line in `Caddyfile`:
```
acme_ca https://acme-staging-v02.api.letsencrypt.org/directory
```

**Note:** Staging certificates will show browser warnings. Remove this line for production.

## Updating BotFather

Once running, update your bot's Web App URL in @BotFather:
```
/myapps → Select your bot → Edit Web App URL
https://tma.yourdomain.com
```

## Troubleshooting

### Certificate Issues
```bash
# Remove old certificates and restart
docker-compose down
rm -rf caddy/data/*
docker-compose up -d
```

### Connection Refused
- Ensure Vite dev server is running on port 5173
- Check `host.docker.internal` resolves (macOS/Windows) or use host network mode on Linux

### Domain Not Resolving
```bash
# Test DNS
dig tma.yourdomain.com
# Should point to your VPS IP
```

## Architecture

```
Internet → Caddy (Docker) → host.docker.internal:5173 → Vite Dev Server
           ↓
        Let's Encrypt
        (Auto HTTPS)
```

## Files

- `docker-compose.yml` - Docker Compose configuration
- `Caddyfile` - Caddy reverse proxy configuration
- `.env` - Environment variables (create from .env.example)
- `caddy/data/` - Caddy data directory (certificates, etc.)
- `caddy/config/` - Caddy configuration directory

## Security Notes

- Let's Encrypt rate limits: 50 certificates per domain per week
- Use staging server for testing
- Certificates auto-renew 30 days before expiration
- Access logs stored in `/data/access.log` inside container
