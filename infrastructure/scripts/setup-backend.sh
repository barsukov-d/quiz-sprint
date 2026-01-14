#!/bin/bash
set -e

echo "🚀 Setting up Quiz Sprint Backend on VPS"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (use sudo)${NC}"
    exit 1
fi

echo -e "${BLUE}📦 Installing Docker and Docker Compose...${NC}"
apt-get update
apt-get install -y ca-certificates curl gnupg lsb-release

# Add Docker's official GPG key
mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Set up Docker repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

echo -e "${GREEN}✅ Docker installed${NC}"

echo -e "${BLUE}👤 Creating quiz-sprint user...${NC}"
if ! id -u quiz-sprint > /dev/null 2>&1; then
    useradd -r -s /bin/bash -d /opt/quiz-sprint quiz-sprint
    echo -e "${GREEN}✅ User created${NC}"
else
    echo -e "${GREEN}✅ User already exists${NC}"
fi

echo -e "${BLUE}📁 Creating directories...${NC}"
mkdir -p /opt/quiz-sprint/{staging,production}
chown -R quiz-sprint:quiz-sprint /opt/quiz-sprint

echo -e "${GREEN}✅ Directories created${NC}"

echo -e "${BLUE}🔧 Installing systemd services...${NC}"
cd "$(dirname "$0")/../systemd"

# Copy systemd service files
cp quiz-sprint-api-staging.service /etc/systemd/system/
cp quiz-sprint-api-production.service /etc/systemd/system/

# Reload systemd
systemctl daemon-reload

# Enable services (but don't start yet - wait for first deployment)
systemctl enable quiz-sprint-api-staging
systemctl enable quiz-sprint-api-production

echo -e "${GREEN}✅ Systemd services installed${NC}"

echo -e "${BLUE}🔧 Setting up PostgreSQL with Docker...${NC}"
cd /opt/quiz-sprint

# Create docker-compose.yml
cat > docker-compose.yml <<'EOF'
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: quiz-sprint-postgres
    restart: always
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${POSTGRES_ROOT_PASSWORD}
      POSTGRES_DB: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    networks:
      - quiz-sprint-network

  redis:
    image: redis:7-alpine
    container_name: quiz-sprint-redis
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - quiz-sprint-network

volumes:
  postgres_data:
  redis_data:

networks:
  quiz-sprint-network:
    driver: bridge
EOF

# Create init-db.sql
cat > init-db.sql <<'EOF'
-- Create databases
CREATE DATABASE quiz_sprint_staging;
CREATE DATABASE quiz_sprint_production;

-- Create users
CREATE USER quiz_user_staging WITH PASSWORD 'staging_password_change_me';
CREATE USER quiz_user_production WITH PASSWORD 'production_password_change_me';

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE quiz_sprint_staging TO quiz_user_staging;
GRANT ALL PRIVILEGES ON DATABASE quiz_sprint_production TO quiz_user_production;
EOF

# Create .env for Docker Compose
cat > .env <<'EOF'
POSTGRES_ROOT_PASSWORD=change_me_root_password
EOF

echo -e "${BLUE}⚠️  IMPORTANT: Edit /opt/quiz-sprint/.env and change database passwords${NC}"
echo -e "${BLUE}⚠️  Then run: cd /opt/quiz-sprint && docker compose up -d${NC}"

echo -e "${GREEN}✅ PostgreSQL setup complete${NC}"

echo -e "${BLUE}🔧 Updating nginx configuration...${NC}"
cd "$(dirname "$0")/../nginx/sites-available"

# Copy nginx configs
cp staging.conf /etc/nginx/sites-available/staging.quiz-sprint-tma.online
cp production.conf /etc/nginx/sites-available/production.quiz-sprint-tma.online

# Test nginx config
nginx -t

# Reload nginx
systemctl reload nginx

echo -e "${GREEN}✅ Nginx configuration updated${NC}"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}🎉 Backend setup complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "1. Edit /opt/quiz-sprint/.env and change database passwords"
echo "2. Start databases: cd /opt/quiz-sprint && docker compose up -d"
echo "3. Update GitHub Secrets:"
echo "   - STAGING_DB_USER=quiz_user_staging"
echo "   - STAGING_DB_PASSWORD=<staging_password>"
echo "   - PROD_DB_USER=quiz_user_production"
echo "   - PROD_DB_PASSWORD=<production_password>"
echo "4. Deploy backend using GitHub Actions"
echo "5. Check status:"
echo "   - systemctl status quiz-sprint-api-staging"
echo "   - systemctl status quiz-sprint-api-production"
echo "   - docker compose ps"
echo ""
