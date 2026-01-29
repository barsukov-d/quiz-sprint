#!/bin/bash

# ========================================
# Reset Daily Challenge for Debugging
# ========================================
# Deletes today's daily quiz and all games
# Usage: ./scripts/reset-daily-challenge.sh [DATE]
#   DATE - optional, format YYYY-MM-DD (defaults to today)

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Database connection
DB_USER="quiz_user"
DB_NAME="quiz_sprint_dev"
DB_CONTAINER="quiz-sprint-postgres-dev"

# Get date (default to today)
DATE=${1:-$(date +%Y-%m-%d)}

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Daily Challenge Reset Script${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "Date: ${GREEN}${DATE}${NC}"
echo ""

# Confirm deletion
read -p "This will DELETE all daily games and quiz for ${DATE}. Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    echo -e "${YELLOW}Cancelled.${NC}"
    exit 0
fi

echo ""
echo -e "${YELLOW}Connecting to database...${NC}"

# Delete games for the date
echo -e "${YELLOW}Deleting daily games for ${DATE}...${NC}"
GAMES_DELETED=$(docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -t -c \
    "WITH deleted AS (DELETE FROM daily_games WHERE date = '${DATE}' RETURNING *) SELECT COUNT(*) FROM deleted;" | tr -d ' ')

echo -e "${GREEN}✓ Deleted ${GAMES_DELETED} games${NC}"

# Delete daily quiz for the date
echo -e "${YELLOW}Deleting daily quiz for ${DATE}...${NC}"
QUIZ_DELETED=$(docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -t -c \
    "WITH deleted AS (DELETE FROM daily_quizzes WHERE date = '${DATE}' RETURNING *) SELECT COUNT(*) FROM deleted;" | tr -d ' ')

echo -e "${GREEN}✓ Deleted ${QUIZ_DELETED} quiz${NC}"

# Show stats
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Reset Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "Games deleted:  ${GAMES_DELETED}"
echo -e "Quizzes deleted: ${QUIZ_DELETED}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "1. Refresh your app"
echo -e "2. Start a new Daily Challenge"
echo ""
