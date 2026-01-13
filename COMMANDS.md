# Development Commands Reference

Quick reference for all common development tasks.

## Daily Development Workflow

### Start Development Session

**Terminal 1 - Start Vite Dev Server:**
```bash
cd ~/projects/quiz-sprint/tma
pnpm dev
```

**Terminal 2 - Start SSH Tunnel:**
```bash
cd ~/projects/quiz-sprint
./tunnel-to-vps.sh
```

**Access your TMA:**
```
https://quiz-sprint-tma.online
```

### Stop Development Session

1. Press `Ctrl+C` in tunnel terminal (Terminal 2)
2. Press `Ctrl+C` in Vite terminal (Terminal 1)

---

## Reverse Proxy Deployment

### Option 1: Automatic Deployment (via Git/GitHub Actions)

```bash
cd ~/projects/quiz-sprint

# Make changes to reverse-proxy files
nano reverse-proxy/Caddyfile

# Commit and push (triggers CI/CD)
git add reverse-proxy/
git commit -m "Update Caddyfile configuration"
git push origin main
```

GitHub Actions will automatically deploy to VPS.

### Option 2: Manual Deployment (Quick Testing)

```bash
cd ~/projects/quiz-sprint

# Deploy changes without git commit
./deploy-reverse-proxy-manual.sh
```

---

## SSH & VPS Access

### Connect to VPS

```bash
ssh root@144.31.199.226
```

### Check Reverse Proxy Status on VPS

```bash
ssh root@144.31.199.226 "cd ~/quiz-sprint/reverse-proxy && docker compose ps"
```

### View Reverse Proxy Logs on VPS

```bash
ssh root@144.31.199.226 "cd ~/quiz-sprint/reverse-proxy && docker compose logs -f caddy"
```

### Restart Reverse Proxy on VPS

```bash
ssh root@144.31.199.226 "cd ~/quiz-sprint/reverse-proxy && docker compose restart"
```

### Stop Reverse Proxy on VPS

```bash
ssh root@144.31.199.226 "cd ~/quiz-sprint/reverse-proxy && docker compose down"
```

### Start Reverse Proxy on VPS

```bash
ssh root@144.31.199.226 "cd ~/quiz-sprint/reverse-proxy && docker compose up -d"
```

---

## Troubleshooting Commands

### Check if Vite is Running (on MacBook)

```bash
curl http://localhost:5173
```

Should return HTML. If not, start Vite.

### Check if Tunnel is Connected (on MacBook)

```bash
# Check SSH process
ps aux | grep "ssh.*5173"
```

Should show the SSH tunnel process. If not, run `./tunnel-to-vps.sh`

### Check if Port 5173 is Listening on VPS

```bash
ssh root@144.31.199.226 "netstat -tuln | grep 5173"
```

Should show port 5173 listening if tunnel is connected.

### Check Caddy is Running on VPS

```bash
ssh root@144.31.199.226 "docker ps | grep caddy"
```

### Test Domain SSL Certificate

```bash
curl -I https://quiz-sprint-tma.online
```

Should return `HTTP/2 200` or `HTTP/2 502` (502 means Caddy works, backend unavailable)

### Check Full Request Flow

```bash
# 1. Check Vite locally
curl http://localhost:5173

# 2. Check tunnel port on VPS
ssh root@144.31.199.226 "curl http://localhost:5173"

# 3. Check domain
curl https://quiz-sprint-tma.online
```

---

## Git & Version Control

### Check Current Branch

```bash
git branch
```

### View Uncommitted Changes

```bash
git status
```

### Create Feature Branch

```bash
git checkout -b feature/my-feature
```

### Commit Changes

```bash
git add .
git commit -m "Description of changes"
```

### Push to GitHub

```bash
git push origin main
```

### View Recent Commits

```bash
git log --oneline -10
```

---

## Project Structure

```
quiz-sprint/
├── .github/workflows/
│   └── deploy-reverse-proxy.yml    # CI/CD workflow
├── reverse-proxy/
│   ├── Caddyfile                   # Reverse proxy config
│   ├── docker-compose.yml          # Docker setup
│   └── deploy.sh                   # Manual deploy script (VPS)
├── tma/
│   ├── src/                        # Vue.js source code
│   ├── vite.config.ts              # Vite configuration
│   └── package.json                # Dependencies
├── tunnel-to-vps.sh                # SSH tunnel script (MacBook)
├── deploy-reverse-proxy-manual.sh  # Manual reverse-proxy deploy (MacBook)
├── .env.tunnel                     # Tunnel configuration (MacBook)
└── TUNNEL-SETUP.md                 # Detailed tunnel documentation
```

---

## Configuration Files

### .env.tunnel (MacBook)

Location: `~/projects/quiz-sprint/.env.tunnel`

```bash
VPS_HOST=144.31.199.226
VPS_USER=root
LOCAL_PORT=5173
REMOTE_PORT=5173
```

### vite.config.ts (MacBook)

Location: `~/projects/quiz-sprint/tma/vite.config.ts`

Important settings:
```typescript
server: {
  host: true,
  allowedHosts: ['.trycloudflare.com', 'localhost', 'quiz-sprint-tma.online'],
}
```

### Caddyfile (VPS)

Location: `/root/quiz-sprint/reverse-proxy/Caddyfile` (on VPS)

```
{
    email barsukov.d@gmail.com
}

{$DOMAIN:localhost} {
    reverse_proxy localhost:{$VITE_PORT:5173}
    encode gzip
    log {
        output file /data/access.log
        format json
    }
}
```

---

## Emergency Procedures

### Tunnel Died - Restart It

```bash
cd ~/projects/quiz-sprint
./tunnel-to-vps.sh
```

### Vite Crashed - Restart It

```bash
cd ~/projects/quiz-sprint/tma
pnpm dev
```

### Reverse Proxy Not Working - Restart It

```bash
ssh root@144.31.199.226 "cd ~/quiz-sprint/reverse-proxy && docker compose restart"
```

### Can't SSH to VPS

1. Check internet connection
2. Check VPS is running (via hosting control panel)
3. Try ping: `ping 144.31.199.226`

### Domain Not Resolving

```bash
# Check DNS
nslookup quiz-sprint-tma.online

# Check if VPS IP matches
dig quiz-sprint-tma.online
```

### Port 5173 Already in Use

```bash
# Find process using port 5173
lsof -i :5173

# Kill the process (replace PID with actual number)
kill -9 PID
```

---

## Useful Aliases (Optional)

Add to `~/.zshrc` or `~/.bashrc`:

```bash
# Quick navigation
alias cdq="cd ~/projects/quiz-sprint"
alias cdt="cd ~/projects/quiz-sprint/tma"
alias cdr="cd ~/projects/quiz-sprint/reverse-proxy"

# Development
alias tma-dev="cd ~/projects/quiz-sprint/tma && pnpm dev"
alias tma-tunnel="cd ~/projects/quiz-sprint && ./tunnel-to-vps.sh"

# VPS management
alias vps-ssh="ssh root@144.31.199.226"
alias vps-logs="ssh root@144.31.199.226 'cd ~/quiz-sprint/reverse-proxy && docker compose logs -f caddy'"
alias vps-status="ssh root@144.31.199.226 'cd ~/quiz-sprint/reverse-proxy && docker compose ps'"

# Git shortcuts
alias gs="git status"
alias ga="git add"
alias gc="git commit -m"
alias gp="git push origin main"
```

After adding, reload shell:
```bash
source ~/.zshrc  # or source ~/.bashrc
```

---

## Monitoring & Logs

### View GitHub Actions Run

1. Go to: https://github.com/barsukov-d/quiz-sprint/actions
2. Click on latest workflow run
3. View deployment logs

### View Local Vite Logs

Already visible in Terminal 1 where Vite is running.

### View Tunnel Connection Status

Already visible in Terminal 2 where tunnel is running.

### View Caddy Access Logs on VPS

```bash
ssh root@144.31.199.226 "cd ~/quiz-sprint/reverse-proxy && docker compose exec caddy cat /data/access.log | tail -50"
```

### View Caddy Error Logs on VPS

```bash
ssh root@144.31.199.226 "cd ~/quiz-sprint/reverse-proxy && docker compose logs caddy | tail -50"
```

---

## Tips & Best Practices

1. **Always start tunnel AFTER Vite is running**
2. **Keep tunnel terminal open** - if it closes, tunnel dies
3. **Use tmux** for persistent sessions (tunnel won't die if terminal closes)
4. **Test locally first** - verify `http://localhost:5173` works before troubleshooting tunnel
5. **Check GitHub Actions** - view deployment status before debugging VPS
6. **Commit often** - small commits are easier to debug and rollback
7. **Use feature branches** - for experimental changes

---

## Support & Documentation

- **Tunnel Setup Guide**: `TUNNEL-SETUP.md`
- **Project Issues**: Check `.beads/beads.db` with `sqlite3`
- **GitHub Actions**: https://github.com/barsukov-d/quiz-sprint/actions
- **Vite Docs**: https://vitejs.dev/
- **Caddy Docs**: https://caddyserver.com/docs/
