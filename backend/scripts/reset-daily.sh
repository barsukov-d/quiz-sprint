#!/bin/bash

# ==========================================
# Daily Challenge Reset Script
# ==========================================
# Usage:
#   ./scripts/reset-daily.sh              # Reset your user for today
#   ./scripts/reset-daily.sh USER_ID      # Reset specific user for today
#   ./scripts/reset-daily.sh USER_ID DATE # Reset specific user for specific date
#   ./scripts/reset-daily.sh --all        # Reset all users for today
#   ./scripts/reset-daily.sh --all DATE   # Reset all users for specific date
#   ./scripts/reset-daily.sh --help       # Show help

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DEFAULT_USER_ID="1121083057"
TODAY=$(date +%Y-%m-%d)

# Help message
show_help() {
    echo -e "${BLUE}Daily Challenge Reset Script${NC}"
    echo ""
    echo "Usage:"
    echo "  ./scripts/reset-daily.sh [USER_ID] [DATE]"
    echo ""
    echo "Options:"
    echo "  --all              Reset all users"
    echo "  --help             Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./scripts/reset-daily.sh"
    echo "    → Reset user $DEFAULT_USER_ID for today ($TODAY)"
    echo ""
    echo "  ./scripts/reset-daily.sh 123456789"
    echo "    → Reset user 123456789 for today"
    echo ""
    echo "  ./scripts/reset-daily.sh 123456789 2026-01-27"
    echo "    → Reset user 123456789 for 2026-01-27"
    echo ""
    echo "  ./scripts/reset-daily.sh --all"
    echo "    → Reset all users for today"
    echo ""
    echo "  ./scripts/reset-daily.sh --all 2026-01-27"
    echo "    → Reset all users for 2026-01-27"
    echo ""
}

# Parse arguments
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    show_help
    exit 0
fi

if [ "$1" = "--all" ]; then
    RESET_ALL=true
    USER_ID=""
    DATE=${2:-$TODAY}
else
    RESET_ALL=false
    USER_ID=${1:-$DEFAULT_USER_ID}
    DATE=${2:-$TODAY}
fi

# ==========================================
# Functions
# ==========================================

run_sql() {
    docker compose -f docker-compose.dev.yml exec -T postgres psql -U quiz_user -d quiz_sprint_dev -c "$1"
}

show_current_state() {
    echo -e "${BLUE}📊 Current State:${NC}"
    if [ "$RESET_ALL" = true ]; then
        run_sql "SELECT player_id, date, status, current_streak, (session_state->>'base_score')::int as score FROM daily_games WHERE date = '$DATE' ORDER BY player_id;"
    else
        run_sql "SELECT id, date, status, current_streak, best_streak, (session_state->>'base_score')::int as score FROM daily_games WHERE player_id = '$USER_ID' AND date = '$DATE';"
    fi
}

delete_games() {
    if [ "$RESET_ALL" = true ]; then
        echo -e "${YELLOW}🗑️  Deleting all games for date: $DATE${NC}"
        run_sql "DELETE FROM daily_games WHERE date = '$DATE';"
    else
        echo -e "${YELLOW}🗑️  Deleting game for user $USER_ID on $DATE${NC}"
        run_sql "DELETE FROM daily_games WHERE player_id = '$USER_ID' AND date = '$DATE';"
    fi
}

show_stats() {
    echo -e "${BLUE}📈 Statistics:${NC}"
    if [ "$RESET_ALL" = true ]; then
        run_sql "SELECT COUNT(*) as total_games, COUNT(DISTINCT player_id) as unique_players FROM daily_games WHERE date = '$DATE';"
    else
        run_sql "SELECT COUNT(*) as total_games, MAX(best_streak) as best_streak FROM daily_games WHERE player_id = '$USER_ID';"
    fi
}

# ==========================================
# Main Script
# ==========================================

echo ""
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Daily Challenge Reset Tool          ║${NC}"
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo ""

if [ "$RESET_ALL" = true ]; then
    echo -e "${YELLOW}⚠️  WARNING: Resetting ALL users for $DATE${NC}"
else
    echo -e "Target: ${GREEN}User $USER_ID${NC} on ${GREEN}$DATE${NC}"
fi
echo ""

# Show current state
show_current_state
echo ""

# Confirm deletion
echo -e "${YELLOW}Do you want to proceed with deletion? (y/n)${NC}"
read -r confirmation

if [ "$confirmation" != "y" ] && [ "$confirmation" != "Y" ]; then
    echo -e "${RED}❌ Cancelled${NC}"
    exit 0
fi

echo ""

# Delete games
delete_games

echo ""
echo -e "${GREEN}✅ Reset complete!${NC}"
echo ""

# Show stats
show_stats

echo ""
echo -e "${GREEN}🎮 You can now start a fresh Daily Challenge!${NC}"
echo ""
