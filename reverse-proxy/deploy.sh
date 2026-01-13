#!/bin/bash

# Deployment script for reverse-proxy on VPS
# This script can be run manually or via CI/CD

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Starting reverse-proxy deployment...${NC}"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    exit 1
fi

# Check if Docker Compose is installed
if ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    exit 1
fi

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  .env file not found${NC}"
    echo -e "${YELLOW}Creating .env from .env.example...${NC}"
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${YELLOW}Please edit .env with your domain and settings${NC}"
        exit 1
    else
        echo -e "${RED}❌ .env.example not found${NC}"
        exit 1
    fi
fi

# Load environment variables
source .env

# Validate required variables
if [ -z "$DOMAIN" ]; then
    echo -e "${RED}❌ DOMAIN not set in .env${NC}"
    exit 1
fi

echo -e "${GREEN}📦 Pulling latest Caddy image...${NC}"
docker compose pull

echo -e "${GREEN}🔄 Deploying containers...${NC}"
docker compose up -d --remove-orphans

echo -e "${GREEN}🧹 Cleaning up old images...${NC}"
docker image prune -f

echo -e "${GREEN}📊 Container status:${NC}"
docker compose ps

echo -e "${GREEN}📝 Viewing logs (last 20 lines):${NC}"
docker compose logs --tail=20 caddy

echo ""
echo -e "${GREEN}✅ Deployment complete!${NC}"
echo -e "${GREEN}🌐 Your reverse proxy should be available at: https://${DOMAIN}${NC}"
echo ""
echo -e "${YELLOW}To view logs: docker compose logs -f caddy${NC}"
echo -e "${YELLOW}To stop: docker compose down${NC}"
echo -e "${YELLOW}To restart: docker compose restart${NC}"
