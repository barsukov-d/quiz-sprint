#!/bin/bash

# ==========================================
# Daily Challenge Debug Inspector
# ==========================================
# Show detailed state of daily challenges

set -e

USER_ID=${1:-1121083057}
DATE=${2:-$(date +%Y-%m-%d)}

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

run_sql() {
    docker compose -f docker-compose.dev.yml exec -T postgres psql -U quiz_user -d quiz_sprint_dev -c "$1"
}

echo ""
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Daily Challenge Debug Inspector     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# 1. User's games for date
echo -e "${BLUE}📅 Games for User $USER_ID on $DATE:${NC}"
run_sql "
SELECT
    id,
    status,
    current_streak,
    best_streak,
    (session_state->>'base_score')::int as base_score,
    (session_state->>'current_question_index')::int as question_idx,
    chest_type,
    chest_coins,
    attempt_number
FROM daily_games
WHERE player_id = '$USER_ID' AND date = '$DATE'
ORDER BY attempt_number;
"
echo ""

# 2. User's streak history
echo -e "${BLUE}🔥 Streak History (Last 7 days):${NC}"
run_sql "
SELECT
    date,
    status,
    current_streak,
    best_streak,
    (session_state->>'base_score')::int as score,
    (session_state->>'correct_answers')::int as correct
FROM daily_games
WHERE player_id = '$USER_ID'
ORDER BY date DESC
LIMIT 7;
"
echo ""

# 3. Today's leaderboard
echo -e "${BLUE}🏆 Today's Leaderboard (Top 10):${NC}"
run_sql "
SELECT
    player_id,
    (session_state->>'base_score')::int as score,
    current_streak,
    chest_type,
    (session_state->>'correct_answers')::int as correct
FROM daily_games
WHERE date = '$DATE' AND status = 'completed'
ORDER BY (session_state->>'base_score')::int DESC
LIMIT 10;
"
echo ""

# 4. Daily quiz info
echo -e "${BLUE}📝 Daily Quiz for $DATE:${NC}"
run_sql "
SELECT
    id,
    date,
    jsonb_array_length(question_ids) as num_questions,
    to_timestamp(expires_at) as expires_at
FROM daily_quizzes
WHERE date = '$DATE';
"
echo ""

# 5. Statistics
echo -e "${BLUE}📊 Statistics for $DATE:${NC}"
run_sql "
SELECT
    COUNT(*) as total_games,
    COUNT(DISTINCT player_id) as unique_players,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
    COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress,
    ROUND(AVG(CASE WHEN status = 'completed' THEN (session_state->>'base_score')::int END), 2) as avg_score,
    MAX((session_state->>'base_score')::int) as max_score
FROM daily_games
WHERE date = '$DATE';
"
echo ""

# 6. Check if user can play
echo -e "${BLUE}🎮 Can Play Status:${NC}"
GAME_COUNT=$(run_sql "SELECT COUNT(*) FROM daily_games WHERE player_id = '$USER_ID' AND date = '$DATE';" | grep -E "^ *[0-9]" | tr -d ' ')
if [ "$GAME_COUNT" -eq "0" ]; then
    echo -e "${GREEN}✅ User can start Daily Challenge${NC}"
elif [ "$GAME_COUNT" -eq "1" ]; then
    STATUS=$(run_sql "SELECT status FROM daily_games WHERE player_id = '$USER_ID' AND date = '$DATE' LIMIT 1;" | grep -v status | grep -v row | tr -d ' ')
    if [ "$STATUS" = "in_progress" ]; then
        echo -e "${YELLOW}⏸️  Game in progress - can continue${NC}"
    else
        echo -e "${YELLOW}✅ Completed - can retry (costs 100 coins or ad)${NC}"
    fi
else
    echo -e "${YELLOW}📊 Multiple attempts: $GAME_COUNT${NC}"
fi
echo ""

# 7. Actions available
echo -e "${BLUE}🛠️  Available Actions:${NC}"
echo "  ./scripts/reset-daily.sh $USER_ID $DATE    - Reset this game (with confirmation)"
echo "  ./scripts/quick-reset.sh $USER_ID $DATE    - Quick reset (no confirmation)"
echo "  ./scripts/reset-daily.sh --all $DATE       - Reset all users for this date"
echo ""
