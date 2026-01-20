# Development Tunnel Setup

SSH tunnels for HTTPS TMA (Telegram Mini App) development.

## Why Tunnels?

Telegram Mini Apps require HTTPS to work. These SSH tunnels expose your local development environment (localhost) to your VPS with SSL certificates, allowing you to develop with live reload while testing in Telegram.

## Architecture

```
┌──────────────┐         SSH Tunnel          ┌──────────────┐         nginx         ┌──────────────────┐
│   MacBook    │ ◄───────────────────────────┤     VPS      │ ◄─────────────────────┤    Telegram      │
│              │                               │              │                       │    Web App       │
│ Backend:3000 │ ─────── Reverse ─────────────► :3000        │                       │                  │
│ Frontend:5173│ ─────── Tunnels ─────────────► :5173        │                       │ https://dev.quiz-│
└──────────────┘                               │              │                       │ sprint-tma.online│
                                                │ nginx proxies:                      │                  │
                                                │ /api/* → :3000                      │                  │
                                                │ /ws/* → :3000                       │                  │
                                                │ /* → :5173                          │                  │
                                                └──────────────┘                       └──────────────────┘
```

## Setup

### 1. Start Local Services

**Backend:**
```bash
cd backend
go run cmd/api/main.go
# Runs on http://localhost:3000
```

**Frontend:**
```bash
cd tma
pnpm dev
# Runs on http://localhost:5173
```

### 2. Start SSH Tunnels

**Backend Tunnel (Terminal 1):**
```bash
./dev-tunnel/start-backend-tunnel.sh
# Forwards VPS:3000 → localhost:3000
```

**Frontend Tunnel (Terminal 2):**
```bash
./dev-tunnel/start-frontend-tunnel.sh
# Forwards VPS:5173 → localhost:5173
```

### 3. Access TMA

Open in Telegram or browser:
- **TMA URL**: https://dev.quiz-sprint-tma.online
- **API**: https://dev.quiz-sprint-tma.online/api/v1/quiz
- **WebSocket**: wss://dev.quiz-sprint-tma.online/ws/leaderboard/:id

## How It Works

### SSH Reverse Tunnels
The tunnels use SSH reverse port forwarding (`-R` flag):
- `ssh -R 3000:localhost:3000` - VPS port 3000 → MacBook port 3000
- `ssh -R 5173:localhost:5173` - VPS port 5173 → MacBook port 5173

### nginx Reverse Proxy
nginx on VPS routes requests:
```nginx
# Frontend (Vite dev server with HMR)
location / {
    proxy_pass http://127.0.0.1:5173;
    # WebSocket support for hot module reload
}

# Backend API
location /api/ {
    proxy_pass http://127.0.0.1:3000/api/;
}

# WebSocket for leaderboard
location /ws/ {
    proxy_pass http://127.0.0.1:3000/ws/;
}
```

### Runtime API Detection
The frontend client (`tma/src/api/client.ts`) automatically detects the hostname:
```typescript
if (window.location.hostname === 'dev.quiz-sprint-tma.online') {
    return 'https://dev.quiz-sprint-tma.online/api/v1'
}
```

## Troubleshooting

### Tunnel Disconnected
SSH tunnels may disconnect after inactivity. Restart the script:
```bash
# Ctrl+C to stop
./dev-tunnel/start-backend-tunnel.sh  # Restart
```

### Port Already in Use
If you see "address already in use" on VPS:
```bash
ssh root@144.31.199.226 "pkill -f 'sshd.*3000' && pkill -f 'sshd.*5173'"
```

### Backend Not Responding
Check if backend is running locally:
```bash
curl http://localhost:3000/api/v1/quiz
```

### Frontend Not Loading
Check if Vite dev server is running:
```bash
curl http://localhost:5173
```

### nginx Configuration
Verify nginx config on VPS:
```bash
ssh root@144.31.199.226 "nginx -t && systemctl reload nginx"
```

Check nginx logs:
```bash
ssh root@144.31.199.226 "tail -f /var/log/nginx/dev-tma-error.log"
```

## Files

- `start-backend-tunnel.sh` - Backend API tunnel (port 3000)
- `start-frontend-tunnel.sh` - Frontend Vite dev server tunnel (port 5173)
- `README.md` - This file

## Related Configuration

- **nginx**: `/etc/nginx/sites-available/dev-tma` on VPS
- **SSL Certs**: `/etc/letsencrypt/live/dev.quiz-sprint-tma.online/` on VPS
- **Frontend Client**: `tma/src/api/client.ts` (runtime hostname detection)
- **Environment**: `tma/.env.development` (localhost URLs for local-only dev)

## Notes

- Tunnels keep connections alive with `ServerAliveInterval=60`
- HMR (Hot Module Reload) works through the tunnel
- Changes to code are reflected immediately without redeploying
- Press `Ctrl+C` to stop a tunnel
