# Backend Deployment Guide

Complete guide for deploying Quiz Sprint Backend API to VPS.

## Prerequisites

- VPS with Ubuntu 20.04+ (2GB RAM minimum)
- Root SSH access
- Domain configured with DNS A records

## First-Time Setup

### 1. Prepare VPS

```bash
# SSH into VPS
ssh root@your-vps-ip

# Update system
apt update && apt upgrade -y
```

### 2. Clone Repository

```bash
git clone https://github.com/your-username/quiz-sprint.git
cd quiz-sprint
```

### 3. Run Backend Setup Script

```bash
cd infrastructure/scripts
chmod +x setup-backend.sh
sudo ./setup-backend.sh
```

This script will:
- ✅ Install Docker and Docker Compose
- ✅ Create `quiz-sprint` user
- ✅ Create directories: `/opt/quiz-sprint/{staging,production}`
- ✅ Install systemd services
- ✅ Set up PostgreSQL and Redis via Docker Compose
- ✅ Update nginx configuration

### 4. Configure Database Passwords

```bash
# Edit database passwords
nano /opt/quiz-sprint/.env

# Change these values:
POSTGRES_ROOT_PASSWORD=<strong-password>
```

Edit `/opt/quiz-sprint/init-db.sql` and change:
```sql
CREATE USER quiz_user_staging WITH PASSWORD 'your-staging-password';
CREATE USER quiz_user_production WITH PASSWORD 'your-production-password';
```

### 5. Start Database Containers

```bash
cd /opt/quiz-sprint
docker compose up -d

# Verify containers are running
docker compose ps
```

Expected output:
```
NAME                    STATUS              PORTS
quiz-sprint-postgres    Up                  0.0.0.0:5432->5432/tcp
quiz-sprint-redis       Up                  0.0.0.0:6379->6379/tcp
```

### 6. Configure GitHub Secrets

Add these secrets in GitHub repository settings:

**Existing:**
- `VPS_HOST` - VPS IP address
- `VPS_USER` - SSH user (usually `root`)
- `VPS_SSH_KEY` - Private SSH key
- `TELEGRAM_BOT_TOKEN` - Bot token
- `TELEGRAM_CHAT_ID` - Chat ID for notifications

**New for Backend:**
- `STAGING_DB_USER` → `quiz_user_staging`
- `STAGING_DB_PASSWORD` → Your staging password
- `PROD_DB_USER` → `quiz_user_production`
- `PROD_DB_PASSWORD` → Your production password

### 7. Deploy via GitHub Actions

1. Go to GitHub Actions tab
2. Run "Build Backend" workflow
3. Wait for build to complete
4. Run "Deploy Backend" workflow:
   - Select environment: `staging`
   - Leave artifact empty (uses latest)
5. Wait for deployment to complete

### 8. Verify Deployment

```bash
# Check service status
systemctl status quiz-sprint-api-staging

# Check logs
journalctl -u quiz-sprint-api-staging -f

# Test health endpoint
curl https://staging.quiz-sprint-tma.online/api/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "quiz-sprint-api"
}
```

## Deployment Workflow

### Via GitHub Actions (Recommended)

**Step 1: Build**
```bash
# Automatic on push to main branch with backend changes
# Or manually trigger "Build Backend" workflow
```

**Step 2: Deploy**
```bash
# Manually trigger "Deploy Backend" workflow
# Select environment: staging or production
```

### Manual Deployment

If GitHub Actions is not available:

```bash
# On local machine
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o quiz-sprint-api cmd/api/main.go

# Copy to VPS
scp quiz-sprint-api root@your-vps:/opt/quiz-sprint/staging/

# On VPS
ssh root@your-vps
cd /opt/quiz-sprint/staging
chmod +x quiz-sprint-api
chown quiz-sprint:quiz-sprint quiz-sprint-api
systemctl restart quiz-sprint-api-staging
systemctl status quiz-sprint-api-staging
```

## Service Management

### systemd Commands

```bash
# Start service
systemctl start quiz-sprint-api-staging

# Stop service
systemctl stop quiz-sprint-api-staging

# Restart service
systemctl restart quiz-sprint-api-staging

# Check status
systemctl status quiz-sprint-api-staging

# Enable on boot
systemctl enable quiz-sprint-api-staging

# View logs
journalctl -u quiz-sprint-api-staging -f

# View last 100 lines
journalctl -u quiz-sprint-api-staging -n 100
```

### Docker Compose Commands

```bash
cd /opt/quiz-sprint

# Start containers
docker compose up -d

# Stop containers
docker compose down

# View logs
docker compose logs -f

# Restart specific service
docker compose restart postgres

# Check status
docker compose ps
```

## Monitoring

### Check Service Health

```bash
# Staging
curl https://staging.quiz-sprint-tma.online/api/health

# Production
curl https://quiz-sprint-tma.online/api/health
```

### Check Database Connection

```bash
# Connect to PostgreSQL
docker exec -it quiz-sprint-postgres psql -U postgres

# List databases
\l

# Connect to staging database
\c quiz_sprint_staging

# List tables
\dt

# Exit
\q
```

### Check Logs

```bash
# Backend service logs
journalctl -u quiz-sprint-api-staging -f

# Nginx access logs
tail -f /var/log/nginx/staging-tma-access.log

# Nginx error logs
tail -f /var/log/nginx/staging-tma-error.log

# PostgreSQL logs
docker compose logs postgres -f
```

## Troubleshooting

### Service Won't Start

```bash
# Check service status
systemctl status quiz-sprint-api-staging

# Check logs for errors
journalctl -u quiz-sprint-api-staging -n 50

# Common issues:
# 1. Port already in use
sudo lsof -i :3001

# 2. Missing .env file
ls -la /opt/quiz-sprint/staging/.env

# 3. Wrong permissions
ls -la /opt/quiz-sprint/staging/quiz-sprint-api
chown quiz-sprint:quiz-sprint /opt/quiz-sprint/staging/*
```

### Database Connection Failed

```bash
# Check if PostgreSQL is running
docker compose ps

# Check database credentials in .env
cat /opt/quiz-sprint/staging/.env | grep DB_

# Test connection
docker exec -it quiz-sprint-postgres psql -U quiz_user_staging -d quiz_sprint_staging
```

### 502 Bad Gateway

```bash
# Check if backend service is running
systemctl status quiz-sprint-api-staging

# Check if service is listening on correct port
sudo netstat -tulpn | grep 3001

# Check nginx error log
tail -f /var/log/nginx/staging-tma-error.log

# Test backend directly (bypass nginx)
curl http://localhost:3001/api/health
```

### WebSocket Not Working

```bash
# Check nginx WebSocket proxy configuration
cat /etc/nginx/sites-available/staging.quiz-sprint-tma.online | grep -A 10 "location /ws"

# Test WebSocket connection
wscat -c wss://staging.quiz-sprint-tma.online/ws/leaderboard/<quiz-id>
```

## Rollback

If deployment fails, rollback to previous version:

### Via GitHub Actions

1. Go to "Deploy Backend" workflow
2. Select previous successful artifact
3. Deploy to the environment

### Manual Rollback

```bash
# On VPS
cd /opt/quiz-sprint/staging

# Stop service
systemctl stop quiz-sprint-api-staging

# Restore backup (if you made one)
cp quiz-sprint-api.backup quiz-sprint-api

# Or re-deploy previous version from local machine
# scp old-binary root@vps:/opt/quiz-sprint/staging/quiz-sprint-api

# Start service
systemctl start quiz-sprint-api-staging
```

## Backup

### Database Backup

```bash
# Create backup directory
mkdir -p /opt/quiz-sprint/backups

# Backup staging database
docker exec quiz-sprint-postgres pg_dump -U quiz_user_staging quiz_sprint_staging > /opt/quiz-sprint/backups/staging-$(date +%Y%m%d-%H%M%S).sql

# Backup production database
docker exec quiz-sprint-postgres pg_dump -U quiz_user_production quiz_sprint_production > /opt/quiz-sprint/backups/production-$(date +%Y%m%d-%H%M%S).sql

# Restore from backup
docker exec -i quiz-sprint-postgres psql -U quiz_user_staging quiz_sprint_staging < backup.sql
```

### Binary Backup

```bash
# Before deploying new version
cp /opt/quiz-sprint/staging/quiz-sprint-api /opt/quiz-sprint/staging/quiz-sprint-api.backup
```

## Security

### Firewall

```bash
# Allow only necessary ports
ufw allow 22    # SSH
ufw allow 80    # HTTP
ufw allow 443   # HTTPS
ufw enable

# Block direct access to backend port
# (nginx proxies requests, so no need to expose 3000/3001)
```

### SSL Certificates

```bash
# Renew certificates (automatic via certbot)
certbot renew --dry-run

# Check certificate expiry
certbot certificates
```

### Update System

```bash
# Regular system updates
apt update && apt upgrade -y

# Update Docker images
cd /opt/quiz-sprint
docker compose pull
docker compose up -d
```

## Performance Optimization

### Enable Go Binary Optimizations

Already enabled in GitHub Actions build:
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -o quiz-sprint-api \
  ./cmd/api
```

### PostgreSQL Tuning

Edit `/opt/quiz-sprint/docker-compose.yml`:
```yaml
services:
  postgres:
    command:
      - "postgres"
      - "-c"
      - "max_connections=200"
      - "-c"
      - "shared_buffers=256MB"
      - "-c"
      - "effective_cache_size=1GB"
```

Then restart:
```bash
docker compose down
docker compose up -d
```

## Maintenance

### Regular Tasks

**Daily:**
- Check service status
- Monitor disk space: `df -h`

**Weekly:**
- Check logs for errors
- Review resource usage: `htop`

**Monthly:**
- System updates: `apt update && apt upgrade`
- Database backups
- Review and rotate logs

## Support

For issues, check:
1. Service logs: `journalctl -u quiz-sprint-api-staging`
2. Nginx logs: `/var/log/nginx/`
3. Docker logs: `docker compose logs`
4. GitHub Actions logs

## Next Steps

After successful deployment:
- [ ] Set up monitoring (Prometheus/Grafana)
- [ ] Configure automated backups
- [ ] Set up log rotation
- [ ] Add health check alerts
- [ ] Configure rate limiting
- [ ] Set up CDN for static assets
