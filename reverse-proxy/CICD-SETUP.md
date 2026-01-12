# CI/CD Setup Guide

Complete guide for setting up automated deployment of reverse-proxy to your VPS using GitHub Actions.

## Overview

The CI/CD pipeline automatically deploys the reverse-proxy when you push changes to the `reverse-proxy/` folder on the `main` branch.

**Flow:** Push to GitHub → GitHub Actions → SSH to VPS → Deploy with Docker Compose

---

## Step 1: Prepare VPS Server

### 1.1 Install Docker

SSH into your VPS:

```bash
ssh your-user@your-vps-ip
```

Install Docker:

```bash
# Download and run Docker installation script
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add current user to docker group (to run without sudo)
sudo usermod -aG docker $USER

# Log out and log back in for group changes to take effect
exit
```

SSH back in and verify:

```bash
docker --version
```

### 1.2 Install Docker Compose

```bash
# Install Docker Compose
sudo apt update
sudo apt install docker-compose -y

# Verify installation
docker-compose --version
```

### 1.3 Configure Firewall

Allow HTTP and HTTPS traffic:

```bash
# For UFW (Ubuntu)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 443/udp  # HTTP/3
sudo ufw allow 22/tcp   # SSH (if not already allowed)
sudo ufw enable

# For firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

### 1.4 Setup SSH Key for GitHub Actions

On your local machine, generate an SSH key pair (if you don't have one):

```bash
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/github_actions
```

Copy the public key to your VPS:

```bash
ssh-copy-id -i ~/.ssh/github_actions.pub your-user@your-vps-ip
```

Or manually:

```bash
# On VPS
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# Paste the public key content
nano ~/.ssh/authorized_keys
# Add the content of ~/.ssh/github_actions.pub

chmod 600 ~/.ssh/authorized_keys
```

Test the connection:

```bash
ssh -i ~/.ssh/github_actions your-user@your-vps-ip
```

**Save the private key (`~/.ssh/github_actions`) - you'll need it for GitHub Secrets!**

---

## Step 2: Configure GitHub Repository Secrets

Go to your GitHub repository:

**Settings → Secrets and variables → Actions → New repository secret**

### Required Secrets:

#### 1. `VPS_SSH_KEY`
- **Description:** Private SSH key for VPS access
- **Value:** Copy the entire content of `~/.ssh/github_actions` (private key)

```bash
cat ~/.ssh/github_actions
```

Copy everything including:
```
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----
```

#### 2. `VPS_HOST`
- **Description:** VPS IP address or hostname
- **Value:** Your VPS IP (e.g., `123.45.67.89`) or hostname (e.g., `vps.example.com`)

#### 3. `VPS_USER`
- **Description:** SSH username for VPS
- **Value:** Your SSH username (e.g., `root`, `ubuntu`, `deploy`)

#### 4. `DOMAIN`
- **Description:** Domain name for your TMA
- **Value:** Your domain (e.g., `tma.yourdomain.com`)
- **Note:** Make sure this domain points to your VPS IP via DNS A record

#### 5. `VITE_PORT`
- **Description:** Port where Vite dev server runs
- **Value:** `5173` (default)

---

## Step 3: Configure DNS

Point your domain to your VPS:

1. Go to your domain registrar or DNS provider
2. Add an **A record**:
   - **Host:** `tma` (or `@` for root domain)
   - **Type:** A
   - **Value:** Your VPS IP address
   - **TTL:** 3600 (or default)

Verify DNS propagation:

```bash
dig tma.yourdomain.com
# or
nslookup tma.yourdomain.com
```

---

## Step 4: Update Caddyfile Email

Edit `reverse-proxy/Caddyfile` and replace the email:

```caddyfile
{
    email your-actual-email@example.com
}
```

This email is used for Let's Encrypt certificate notifications.

---

## Step 5: Test Deployment

### Manual Test (Optional)

Test the deployment manually first:

```bash
# On VPS
cd ~/quiz-sprint/reverse-proxy

# Copy files manually for testing
# ... (files should be copied here)

# Create .env
cat > .env << EOF
DOMAIN=tma.yourdomain.com
VITE_PORT=5173
EOF

# Run deployment
./deploy.sh
```

### Automated Deployment

Commit and push changes:

```bash
git add .
git commit -m "Setup CI/CD for reverse-proxy"
git push origin main
```

Monitor deployment:

1. Go to GitHub → **Actions** tab
2. Click on the latest workflow run
3. Watch the deployment progress

---

## Step 6: Verify Deployment

After successful deployment:

### Check Container Status

SSH into VPS:

```bash
ssh your-user@your-vps-ip
cd ~/quiz-sprint/reverse-proxy
docker-compose ps
```

You should see:

```
       Name                     Command               State                    Ports
--------------------------------------------------------------------------------------------------------
tma-reverse-proxy   caddy run --config /etc/ca ...   Up      0.0.0.0:443->443/tcp, 0.0.0.0:80->80/tcp
```

### View Logs

```bash
docker-compose logs -f caddy
```

### Test HTTPS

```bash
curl -I https://tma.yourdomain.com
```

Should return `200 OK` (or `502 Bad Gateway` if Vite server is not running yet).

### Check Certificate

```bash
echo | openssl s_client -servername tma.yourdomain.com -connect tma.yourdomain.com:443 | grep subject
```

Should show Let's Encrypt certificate.

---

## Workflow Triggers

The GitHub Actions workflow runs when:

1. **Push to main branch** with changes in:
   - `reverse-proxy/**`
   - `.github/workflows/deploy-reverse-proxy.yml`

2. **Manual trigger** via GitHub Actions UI:
   - Go to Actions → "Deploy Reverse Proxy to VPS" → Run workflow

---

## Troubleshooting

### SSH Connection Failed

```
Error: Permission denied (publickey)
```

**Solution:**
- Verify `VPS_SSH_KEY` secret contains the correct private key
- Ensure public key is in `~/.ssh/authorized_keys` on VPS
- Check SSH permissions: `chmod 600 ~/.ssh/authorized_keys`

### Docker Not Found

```
Error: docker: command not found
```

**Solution:**
- Install Docker on VPS (see Step 1.1)
- Verify user is in docker group: `groups`

### Certificate Error

```
Error: obtaining certificate
```

**Solution:**
- Verify domain DNS points to VPS IP
- Check firewall allows ports 80 and 443
- Use Let's Encrypt staging for testing (edit Caddyfile)

### Health Check Failed

```
Error: curl failed
```

**Solution:**
- Ensure Vite dev server is running on VPS
- Check if Caddy container is running: `docker-compose ps`
- View Caddy logs: `docker-compose logs caddy`

---

## Security Best Practices

1. **Use dedicated SSH key** for GitHub Actions (not your personal key)
2. **Restrict SSH key permissions** on VPS: `chmod 600 ~/.ssh/authorized_keys`
3. **Use non-root user** for deployments
4. **Enable UFW/firewall** and only allow necessary ports
5. **Regularly rotate secrets** (SSH keys, etc.)
6. **Use GitHub Environments** for production deployments (optional)

---

## Maintenance

### Update Caddy Image

```bash
docker-compose pull
docker-compose up -d
```

### View Logs

```bash
docker-compose logs -f caddy
```

### Stop Proxy

```bash
docker-compose down
```

### Restart Proxy

```bash
docker-compose restart
```

### Clean Up

```bash
# Remove old images
docker image prune -f

# Remove all unused resources
docker system prune -a -f
```

---

## Next Steps

After CI/CD is working:

1. ✅ Deploy TMA application to VPS
2. ✅ Update @BotFather with your HTTPS URL
3. ✅ Test TMA in Telegram
4. ✅ Monitor logs and performance

---

## Support

For issues:
- Check GitHub Actions logs
- Review VPS logs: `docker-compose logs caddy`
- Verify DNS and firewall settings
- Check Beads tasks: `bd list`
