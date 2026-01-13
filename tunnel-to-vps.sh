#!/bin/bash

# SSH Reverse Tunnel Script
# Forwards local Vite dev server (localhost:5173) to VPS
# So Caddy on VPS can proxy requests to your MacBook

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Load .env.tunnel if exists
if [ -f .env.tunnel ]; then
    echo -e "${GREEN}📝 Loading configuration from .env.tunnel${NC}"
    export $(cat .env.tunnel | grep -v '^#' | xargs)
fi

# Configuration (with defaults)
VPS_HOST="${VPS_HOST:-}"
VPS_USER="${VPS_USER:-deploy}"
LOCAL_PORT="${LOCAL_PORT:-5173}"
REMOTE_PORT="${REMOTE_PORT:-5173}"

# Check if VPS_HOST is set
if [ -z "$VPS_HOST" ]; then
    echo -e "${RED}❌ VPS_HOST environment variable not set${NC}"
    echo -e "${YELLOW}Usage: VPS_HOST=your-vps-ip ./tunnel-to-vps.sh${NC}"
    echo -e "${YELLOW}Or create .env.tunnel file with VPS_HOST=your-vps-ip${NC}"
    exit 1
fi

echo -e "${GREEN}🚇 Starting SSH reverse tunnel...${NC}"
echo -e "${GREEN}   Local:  localhost:${LOCAL_PORT} (MacBook)${NC}"
echo -e "${GREEN}   Remote: localhost:${REMOTE_PORT} (VPS)${NC}"
echo -e "${GREEN}   VPS:    ${VPS_USER}@${VPS_HOST}${NC}"
echo ""
echo -e "${YELLOW}⚠️  Make sure your Vite dev server is running: pnpm dev${NC}"
echo ""
echo -e "${GREEN}Press Ctrl+C to stop the tunnel${NC}"
echo ""

# Start SSH reverse tunnel with auto-reconnect
while true; do
    ssh -N \
        -R ${REMOTE_PORT}:localhost:${LOCAL_PORT} \
        -o ServerAliveInterval=60 \
        -o ServerAliveCountMax=3 \
        -o ExitOnForwardFailure=yes \
        ${VPS_USER}@${VPS_HOST} || {

        echo -e "${RED}❌ Tunnel disconnected. Reconnecting in 5 seconds...${NC}"
        sleep 5
    }
done
