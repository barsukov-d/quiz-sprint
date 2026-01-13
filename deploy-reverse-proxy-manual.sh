#!/bin/bash

# Manual deployment script for reverse-proxy
# Use this to deploy changes without committing to git

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Load VPS config
if [ -f .env.tunnel ]; then
    export $(cat .env.tunnel | grep -v '^#' | xargs)
fi

VPS_USER="${VPS_USER:-root}"
VPS_HOST="${VPS_HOST:-}"

if [ -z "$VPS_HOST" ]; then
    echo "❌ VPS_HOST not set"
    exit 1
fi

echo -e "${GREEN}📦 Deploying reverse-proxy to VPS...${NC}"

# Copy files to VPS
echo -e "${YELLOW}Copying files...${NC}"
scp -r reverse-proxy/* ${VPS_USER}@${VPS_HOST}:~/quiz-sprint/reverse-proxy/

# Deploy on VPS
echo -e "${YELLOW}Restarting containers...${NC}"
ssh ${VPS_USER}@${VPS_HOST} << 'EOF'
cd ~/quiz-sprint/reverse-proxy
docker compose down
docker compose pull
docker compose up -d
docker compose ps
EOF

echo -e "${GREEN}✅ Deployment complete!${NC}"
