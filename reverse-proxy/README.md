# Reverse Proxy Setup (Development Only)

This reverse proxy is configured **for development purposes only**. It allows you to develop your TMA locally on your MacBook while exposing it via HTTPS for Telegram WebView testing.

## Architecture

```
MacBook (Vite Dev)  →  SSH Tunnel  →  VPS (Caddy)  →  HTTPS Domain
localhost:5173      →  Port 5173    →  quiz-sprint-tma.online
```

## Important Notes

⚠️ **Development Only** - This setup is NOT for production:
- No load balancing
- No redundancy
- Requires active SSH tunnel from your MacBook
- Single point of failure (your MacBook)

✅ **For Development:**
- Hot reload works perfectly
- Fast iteration cycle
- No need to deploy TMA to VPS
- Develop on your local machine

## Daily Usage

### Start Development:
1. On MacBook: Start Vite (`cd ~/projects/quiz-sprint/tma && pnpm dev`)
2. On MacBook: Start tunnel (`cd ~/projects/quiz-sprint && ./tunnel-to-vps.sh`)
3. Access: `https://quiz-sprint-tma.online`

### Stop Development:
1. Press `Ctrl+C` on tunnel terminal
2. Press `Ctrl+C` on Vite terminal

The reverse proxy on VPS keeps running 24/7, ready for your next dev session.

## Configuration

### Caddyfile
- Configured for `localhost:5173` (tunnel endpoint)
- HTTPS automatic via Let's Encrypt
- WebSocket support for Vite HMR
- Gzip compression enabled

### docker-compose.yml
- Uses `network_mode: host` to access localhost services
- Auto-restart enabled (VPS reboots)
- Persistent SSL certificates in `./caddy/data`

## When to Modify

Only modify this reverse proxy when:
- Changing domain
- Adding more dev ports (staging, etc.)
- Adjusting SSL settings
- Adding CORS headers

For TMA code changes, just edit locally and Vite will hot-reload!

## Production Deployment

When ready for production, you'll need:
1. Deploy TMA build to VPS (not dev server)
2. Separate production Caddy config
3. Process manager (PM2, systemd)
4. Monitoring & logging
5. Separate environment (production subdomain/domain)

See issue `quiz-sprint-ahq` in Beads tracker for multi-environment setup.

## Troubleshooting

**502 Bad Gateway:**
- Check tunnel is running: `ps aux | grep "ssh.*5173"`
- Check Vite is running: `curl http://localhost:5173`
- Restart tunnel: `./tunnel-to-vps.sh`

**SSL Certificate Issues:**
- Check Caddy logs: `docker compose logs caddy`
- Verify email in Caddyfile: `barsukov.d@gmail.com`
- Check domain DNS points to VPS IP

**Port Conflicts:**
- Only one dev session at a time (port 5173)
- If port busy: `lsof -i :5173` and kill process

## Files

```
reverse-proxy/
├── Caddyfile              # Reverse proxy config
├── docker-compose.yml     # Docker container config
├── .env.example           # Environment template
├── caddy/
│   ├── data/              # SSL certificates (auto-generated)
│   └── config/            # Caddy config cache
└── README.md             # This file
```

## See Also

- `TUNNEL-SETUP.md` - Full SSH tunnel documentation
- `COMMANDS.md` - Common commands reference
- `.github/workflows/deploy-reverse-proxy.yml` - Auto-deployment config
