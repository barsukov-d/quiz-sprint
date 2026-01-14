#!/bin/bash
set -e

echo "🚀 Setting up Quiz Sprint Backend on VPS (Docker Edition)"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

# ==========================================
# 1. Install Docker
# ==========================================
echo -e "${BLUE}📦 Checking Docker installation...${NC}"

if command -v docker &> /dev/null; then
    echo -e "${GREEN}✅ Docker already installed: $(docker --version)${NC}"
else
    echo -e "${BLUE}📦 Installing Docker...${NC}"
    curl -fsSL https://get.docker.com | sh
    echo -e "${GREEN}✅ Docker installed${NC}"
fi

# Verify Docker Compose plugin
if docker compose version &> /dev/null; then
    echo -e "${GREEN}✅ Docker Compose available: $(docker compose version)${NC}"
else
    echo -e "${RED}❌ Docker Compose not available. Please reinstall Docker.${NC}"
    exit 1
fi

# ==========================================
# 2. Create directories
# ==========================================
echo -e "${BLUE}📁 Creating directories...${NC}"

mkdir -p /opt/quiz-sprint/staging
mkdir -p /opt/quiz-sprint/production

echo -e "${GREEN}✅ Directories created:${NC}"
echo "   - /opt/quiz-sprint/staging"
echo "   - /opt/quiz-sprint/production"

# ==========================================
# 3. Login to GitHub Container Registry
# ==========================================
echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}⚠️  GitHub Container Registry Login${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo "To pull Docker images from ghcr.io, you need to login."
echo ""
echo "Option 1: Use GitHub Personal Access Token (PAT)"
echo "  1. Go to: https://github.com/settings/tokens"
echo "  2. Generate token with 'read:packages' scope"
echo "  3. Run: echo YOUR_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin"
echo ""
echo "Option 2: GitHub Actions will login automatically during deployment"
echo ""

# ==========================================
# 4. Summary
# ==========================================
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}🎉 Backend setup complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}What was installed:${NC}"
echo "  ✅ Docker Engine"
echo "  ✅ Docker Compose"
echo "  ✅ Directories: /opt/quiz-sprint/{staging,production}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo ""
echo "1. Configure GitHub Secrets (if not done):"
echo "   - VPS_HOST=144.31.199.226"
echo "   - VPS_USER=root"
echo "   - VPS_SSH_KEY=<your private key>"
echo "   - STAGING_DB_USER=quiz_user"
echo "   - STAGING_DB_PASSWORD=<your password>"
echo "   - PROD_DB_USER=quiz_user"
echo "   - PROD_DB_PASSWORD=<your password>"
echo ""
echo "2. Run GitHub Actions workflow:"
echo "   - Go to: GitHub → Actions → Deploy Backend (Docker)"
echo "   - Select environment: staging"
echo "   - Run workflow"
echo ""
echo "3. Verify deployment:"
echo "   curl https://staging.quiz-sprint-tma.online/api/health"
echo ""
echo -e "${BLUE}Useful commands:${NC}"
echo ""
echo "  # Check staging containers"
echo "  cd /opt/quiz-sprint/staging && docker compose ps"
echo ""
echo "  # View staging logs"
echo "  cd /opt/quiz-sprint/staging && docker compose logs -f api"
echo ""
echo "  # Restart staging"
echo "  cd /opt/quiz-sprint/staging && docker compose restart"
echo ""
echo "  # Check production containers"
echo "  cd /opt/quiz-sprint/production && docker compose ps"
echo ""
