# SSH Reverse Tunnel Setup for TMA Development

This guide helps you develop your TMA on your MacBook while exposing it through your VPS.

## How It Works

```
MacBook (Vite)  →  SSH Tunnel  →  VPS (Caddy)  →  Internet (HTTPS)
localhost:5173  →  Port Forward  →  quiz-sprint-tma.online
```

## Prerequisites

- ✅ VPS with Caddy reverse proxy (already set up)
- ✅ SSH access to your VPS
- ✅ Vite dev server running on your MacBook

## Setup Instructions

### 1. Configure VPS SSH (One-time setup)

SSH into your VPS and enable `GatewayPorts`:

```bash
ssh deploy@your-vps-ip

# Edit SSH config
sudo nano /etc/ssh/sshd_config

# Add or modify this line:
GatewayPorts clientspecified

# Save and restart SSH
sudo systemctl restart sshd
```

### 2. Configure Tunnel on MacBook

Create `.env.tunnel` file:

```bash
cd ~/projects/quiz-sprint
cp .env.tunnel.example .env.tunnel
nano .env.tunnel
```

Edit with your VPS details:
```bash
VPS_HOST=144.31.199.226  # Your VPS IP
VPS_USER=deploy          # Your VPS user
LOCAL_PORT=5173
REMOTE_PORT=5173
```

### 3. Start Development

**Terminal 1: Start Vite dev server**
```bash
cd ~/projects/quiz-sprint/tma
pnpm dev
```

**Terminal 2: Start SSH tunnel**
```bash
cd ~/projects/quiz-sprint
./tunnel-to-vps.sh
```

You should see:
```
🚇 Starting SSH reverse tunnel...
   Local:  localhost:5173 (MacBook)
   Remote: localhost:5173 (VPS)
   VPS:    deploy@144.31.199.226

⚠️  Make sure your Vite dev server is running: pnpm dev

Press Ctrl+C to stop the tunnel
```

**Terminal 3: (Optional) Monitor logs on VPS**
```bash
ssh deploy@your-vps-ip
cd ~/quiz-sprint/reverse-proxy
docker compose logs -f caddy
```

### 4. Access Your TMA

Open in browser or Telegram:
```
https://quiz-sprint-tma.online
```

## Troubleshooting

### 502 Bad Gateway
- ✅ Check Vite is running: `http://localhost:5173` should work on MacBook
- ✅ Check tunnel is connected: Should show connection in terminal
- ✅ Check Caddy logs: `docker compose logs caddy` on VPS

### Connection Refused
- ✅ Check SSH key is set up for VPS
- ✅ Verify VPS_HOST in `.env.tunnel` is correct
- ✅ Make sure firewall allows SSH (port 22)

### Tunnel Keeps Disconnecting
- ✅ Check internet connection
- ✅ The script auto-reconnects after 5 seconds
- ✅ Check VPS is reachable: `ping your-vps-ip`

### Permission Denied
- ✅ Check SSH key: `ssh deploy@your-vps-ip` should work without password
- ✅ Make sure your public key is in VPS `~/.ssh/authorized_keys`

## Tips

### Auto-start tunnel when Mac starts

Create a launchd service or use a terminal multiplexer like `tmux`:

```bash
# Install tmux
brew install tmux

# Create persistent session
tmux new -s tma-dev

# Start Vite in one pane
pnpm dev

# Split window (Ctrl+B, %)
# Start tunnel in another pane
./tunnel-to-vps.sh

# Detach: Ctrl+B, D
# Reattach: tmux attach -t tma-dev
```

### Quick start script

Create `start-dev.sh`:
```bash
#!/bin/bash
cd ~/projects/quiz-sprint/tma
pnpm dev &
cd ~/projects/quiz-sprint
./tunnel-to-vps.sh
```

## Security Notes

- The tunnel only forwards port 5173 from MacBook to VPS
- Only localhost:5173 is accessible on VPS (not exposed to internet directly)
- Caddy handles HTTPS/SSL certificates
- All traffic is encrypted via SSH tunnel

## Architecture

```
┌──────────────────┐
│   Your MacBook   │
│                  │
│  Vite Dev Server │
│  localhost:5173  │
└────────┬─────────┘
         │ SSH Tunnel
         │ (Encrypted)
         │
┌────────▼─────────┐
│   VPS Server     │
│                  │
│  localhost:5173  │◄──────┐
│        │         │       │
│  ┌─────▼──────┐  │       │
│  │   Caddy    │  │       │
│  │  (Docker)  │  │       │
│  └─────┬──────┘  │       │
│        │         │       │
│   Ports 80/443   │       │
└────────┬─────────┘       │
         │                 │
         │ HTTPS           │
         │                 │
┌────────▼─────────────────┴──┐
│  Internet / Telegram Bot    │
│  quiz-sprint-tma.online     │
└─────────────────────────────┘
```

## Stopping Development

1. Press `Ctrl+C` in tunnel terminal
2. Press `Ctrl+C` in Vite terminal
3. Done! VPS keeps running, ready for next session.
