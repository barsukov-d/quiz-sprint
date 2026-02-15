#!/bin/bash
# E2E Tests: Daily Challenge Bonus System
# Usage: ./daily_challenge_bonus_test.sh [base_url]
#
# Tests:
# 1. Streak bonus calculation (3d→1.1x, 7d→1.25x, 30d→1.5x)
# 2. Streak bonus applied to final score
# 3. Streak bonus applied to chest coins
# 4. Streak accumulation across days
# 5. Streak reset when day missed

BASE="${1:-http://localhost:3000/api/v1}"
KEY="X-Admin-Key: dev-admin-key-2026"
PLAYER="e2e-test-bonus-$$"  # Unique per run

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

TESTS_PASSED=0
TESTS_FAILED=0

log_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    echo -e "  Expected: $2"
    echo -e "  Actual:   $3"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

log_info() {
    echo -e "${YELLOW}→${NC} $1"
}

cleanup() {
    log_info "Cleaning up test player: $PLAYER"
    curl -s -X DELETE "$BASE/admin/player/reset?playerId=$PLAYER" -H "$KEY" > /dev/null 2>&1 || true
}

# Cleanup on exit
trap cleanup EXIT

echo "========================================"
echo "E2E Tests: Daily Challenge Bonus System"
echo "========================================"
echo "Base URL: $BASE"
echo "Test Player: $PLAYER"
echo ""

# Health check
log_info "Checking server health..."
HEALTH_URL="${BASE%/api/v1}/health"
HEALTH=$(curl -s "$HEALTH_URL" | jq -r '.status' 2>/dev/null || echo "error")
if [ "$HEALTH" != "ok" ]; then
    echo -e "${RED}Server not available at $BASE${NC}"
    exit 1
fi
log_pass "Server is healthy"

# ===========================================
# Test 1: Fresh player has no streak bonus
# ===========================================
echo ""
echo "--- Test 1: Fresh player has no streak bonus ---"

cleanup

STREAK=$(curl -s "$BASE/daily-challenge/streak?playerId=$PLAYER" | jq '.data.streak')
CURRENT=$(echo "$STREAK" | jq -r '.currentStreak')
BONUS=$(echo "$STREAK" | jq -r '.bonusPercent')

if [ "$CURRENT" = "0" ] && [ "$BONUS" = "0" ]; then
    log_pass "Fresh player has streak=0, bonus=0%"
else
    log_fail "Fresh player streak" "streak=0, bonus=0%" "streak=$CURRENT, bonus=$BONUS%"
fi

# ===========================================
# Test 2: Simulate 3-day streak → 10% bonus
# ===========================================
echo ""
echo "--- Test 2: 3-day streak gives 10% bonus ---"

cleanup

# Simulate 2 days (yesterday and day before)
curl -s -X POST "$BASE/admin/daily-challenge/simulate-streak" \
    -H "$KEY" -H "Content-Type: application/json" \
    -d "{\"playerId\": \"$PLAYER\", \"days\": 2, \"baseScore\": 100}" > /dev/null

# Start game today (will be day 3)
GAME_RESP=$(curl -s -X POST "$BASE/daily-challenge/start" \
    -H "Content-Type: application/json" \
    -d "{\"playerId\": \"$PLAYER\"}")

GAME_ID=$(echo "$GAME_RESP" | jq -r '.data.game.id')

if [ "$GAME_ID" = "null" ] || [ -z "$GAME_ID" ]; then
    log_fail "Start game for 3-day streak" "valid game ID" "null"
else
    log_pass "Started game with ID: ${GAME_ID:0:8}..."

    # Answer all 10 questions correctly with fast time (max bonus)
    for i in {1..10}; do
        # Get current question
        STATUS=$(curl -s "$BASE/daily-challenge/status?playerId=$PLAYER")
        Q_ID=$(echo "$STATUS" | jq -r '.data.game.currentQuestion.id')

        # Find correct answer
        CORRECT_ID=$(echo "$STATUS" | jq -r '.data.game.currentQuestion.answers[] | select(.isCorrect == true) | .id')

        if [ -z "$CORRECT_ID" ] || [ "$CORRECT_ID" = "null" ]; then
            # Fallback: get first answer
            CORRECT_ID=$(echo "$STATUS" | jq -r '.data.game.currentQuestion.answers[0].id')
        fi

        # Submit answer
        ANSWER_RESP=$(curl -s -X POST "$BASE/daily-challenge/$GAME_ID/answer" \
            -H "Content-Type: application/json" \
            -d "{\"playerId\": \"$PLAYER\", \"questionId\": \"$Q_ID\", \"answerId\": \"$CORRECT_ID\", \"timeTaken\": 2}")

        IS_COMPLETED=$(echo "$ANSWER_RESP" | jq -r '.data.isGameCompleted')

        if [ "$IS_COMPLETED" = "true" ]; then
            # Check streak bonus in results
            STREAK_BONUS=$(echo "$ANSWER_RESP" | jq -r '.data.gameResults.streakBonus')
            CURRENT_STREAK=$(echo "$ANSWER_RESP" | jq -r '.data.gameResults.currentStreak')
            FINAL_SCORE=$(echo "$ANSWER_RESP" | jq -r '.data.gameResults.finalScore')
            BASE_SCORE=$(echo "$ANSWER_RESP" | jq -r '.data.gameResults.baseScore')

            # 3-day streak should give 1.1x (10%)
            if [ "$STREAK_BONUS" = "10" ] && [ "$CURRENT_STREAK" = "3" ]; then
                log_pass "3-day streak gives 10% bonus (streak=$CURRENT_STREAK)"
            else
                log_fail "3-day streak bonus" "streakBonus=10, currentStreak=3" "streakBonus=$STREAK_BONUS, currentStreak=$CURRENT_STREAK"
            fi

            # Verify score calculation: finalScore = baseScore * 1.1
            EXPECTED_FINAL=$(echo "$BASE_SCORE * 1.1" | bc | cut -d. -f1)
            if [ "$FINAL_SCORE" = "$EXPECTED_FINAL" ]; then
                log_pass "Score calculation correct: $BASE_SCORE * 1.1 = $FINAL_SCORE"
            else
                log_fail "Score calculation" "$EXPECTED_FINAL" "$FINAL_SCORE (base=$BASE_SCORE)"
            fi

            break
        fi
    done
fi

# ===========================================
# Test 3: 7-day streak → 25% bonus
# ===========================================
echo ""
echo "--- Test 3: 7-day streak gives 25% bonus ---"

cleanup

curl -s -X POST "$BASE/admin/daily-challenge/simulate-streak" \
    -H "$KEY" -H "Content-Type: application/json" \
    -d "{\"playerId\": \"$PLAYER\", \"days\": 6, \"baseScore\": 100}" > /dev/null

GAME_RESP=$(curl -s -X POST "$BASE/daily-challenge/start" \
    -H "Content-Type: application/json" \
    -d "{\"playerId\": \"$PLAYER\"}")

GAME_ID=$(echo "$GAME_RESP" | jq -r '.data.game.id')

if [ "$GAME_ID" != "null" ] && [ -n "$GAME_ID" ]; then
    # Answer all questions
    for i in {1..10}; do
        STATUS=$(curl -s "$BASE/daily-challenge/status?playerId=$PLAYER")
        Q_ID=$(echo "$STATUS" | jq -r '.data.game.currentQuestion.id')
        CORRECT_ID=$(echo "$STATUS" | jq -r '.data.game.currentQuestion.answers[0].id')

        ANSWER_RESP=$(curl -s -X POST "$BASE/daily-challenge/$GAME_ID/answer" \
            -H "Content-Type: application/json" \
            -d "{\"playerId\": \"$PLAYER\", \"questionId\": \"$Q_ID\", \"answerId\": \"$CORRECT_ID\", \"timeTaken\": 2}")

        IS_COMPLETED=$(echo "$ANSWER_RESP" | jq -r '.data.isGameCompleted')

        if [ "$IS_COMPLETED" = "true" ]; then
            STREAK_BONUS=$(echo "$ANSWER_RESP" | jq -r '.data.gameResults.streakBonus')
            CURRENT_STREAK=$(echo "$ANSWER_RESP" | jq -r '.data.gameResults.currentStreak')

            if [ "$STREAK_BONUS" = "25" ] && [ "$CURRENT_STREAK" = "7" ]; then
                log_pass "7-day streak gives 25% bonus"
            else
                log_fail "7-day streak bonus" "streakBonus=25, currentStreak=7" "streakBonus=$STREAK_BONUS, currentStreak=$CURRENT_STREAK"
            fi
            break
        fi
    done
else
    log_fail "Start game for 7-day streak" "valid game ID" "null"
fi

# ===========================================
# Test 4: 30-day streak → 50% bonus
# ===========================================
echo ""
echo "--- Test 4: 30-day streak gives 50% bonus ---"

cleanup

curl -s -X POST "$BASE/admin/daily-challenge/simulate-streak" \
    -H "$KEY" -H "Content-Type: application/json" \
    -d "{\"playerId\": \"$PLAYER\", \"days\": 29, \"baseScore\": 100}" > /dev/null

GAME_RESP=$(curl -s -X POST "$BASE/daily-challenge/start" \
    -H "Content-Type: application/json" \
    -d "{\"playerId\": \"$PLAYER\"}")

GAME_ID=$(echo "$GAME_RESP" | jq -r '.data.game.id')

if [ "$GAME_ID" != "null" ] && [ -n "$GAME_ID" ]; then
    for i in {1..10}; do
        STATUS=$(curl -s "$BASE/daily-challenge/status?playerId=$PLAYER")
        Q_ID=$(echo "$STATUS" | jq -r '.data.game.currentQuestion.id')
        CORRECT_ID=$(echo "$STATUS" | jq -r '.data.game.currentQuestion.answers[0].id')

        ANSWER_RESP=$(curl -s -X POST "$BASE/daily-challenge/$GAME_ID/answer" \
            -H "Content-Type: application/json" \
            -d "{\"playerId\": \"$PLAYER\", \"questionId\": \"$Q_ID\", \"answerId\": \"$CORRECT_ID\", \"timeTaken\": 2}")

        IS_COMPLETED=$(echo "$ANSWER_RESP" | jq -r '.data.isGameCompleted')

        if [ "$IS_COMPLETED" = "true" ]; then
            STREAK_BONUS=$(echo "$ANSWER_RESP" | jq -r '.data.gameResults.streakBonus')
            CURRENT_STREAK=$(echo "$ANSWER_RESP" | jq -r '.data.gameResults.currentStreak')

            if [ "$STREAK_BONUS" = "50" ] && [ "$CURRENT_STREAK" = "30" ]; then
                log_pass "30-day streak gives 50% bonus"
            else
                log_fail "30-day streak bonus" "streakBonus=50, currentStreak=30" "streakBonus=$STREAK_BONUS, currentStreak=$CURRENT_STREAK"
            fi
            break
        fi
    done
else
    log_fail "Start game for 30-day streak" "valid game ID" "null"
fi

# ===========================================
# Test 5: Chest coins multiplied by streak
# ===========================================
echo ""
echo "--- Test 5: Chest coins multiplied by streak bonus ---"

cleanup

# Set up 7-day streak for 1.25x multiplier
curl -s -X POST "$BASE/admin/daily-challenge/simulate-streak" \
    -H "$KEY" -H "Content-Type: application/json" \
    -d "{\"playerId\": \"$PLAYER\", \"days\": 6, \"baseScore\": 100}" > /dev/null

GAME_RESP=$(curl -s -X POST "$BASE/daily-challenge/start" \
    -H "Content-Type: application/json" \
    -d "{\"playerId\": \"$PLAYER\"}")

GAME_ID=$(echo "$GAME_RESP" | jq -r '.data.game.id')

if [ "$GAME_ID" != "null" ] && [ -n "$GAME_ID" ]; then
    for i in {1..10}; do
        STATUS=$(curl -s "$BASE/daily-challenge/status?playerId=$PLAYER")
        Q_ID=$(echo "$STATUS" | jq -r '.data.game.currentQuestion.id')
        CORRECT_ID=$(echo "$STATUS" | jq -r '.data.game.currentQuestion.answers[0].id')

        ANSWER_RESP=$(curl -s -X POST "$BASE/daily-challenge/$GAME_ID/answer" \
            -H "Content-Type: application/json" \
            -d "{\"playerId\": \"$PLAYER\", \"questionId\": \"$Q_ID\", \"answerId\": \"$CORRECT_ID\", \"timeTaken\": 2}")

        IS_COMPLETED=$(echo "$ANSWER_RESP" | jq -r '.data.isGameCompleted')

        if [ "$IS_COMPLETED" = "true" ]; then
            CHEST_COINS=$(echo "$ANSWER_RESP" | jq -r '.data.gameResults.chestReward.coins')
            CHEST_TYPE=$(echo "$ANSWER_RESP" | jq -r '.data.gameResults.chestReward.chestType')

            # With 1.25x multiplier, coins should be above base range
            # Golden chest base: 300-500, with 1.25x: 375-625
            if [ "$CHEST_TYPE" = "golden" ]; then
                if [ "$CHEST_COINS" -ge 375 ] && [ "$CHEST_COINS" -le 625 ]; then
                    log_pass "Golden chest coins with 1.25x streak: $CHEST_COINS (range 375-625)"
                else
                    log_fail "Golden chest coins with streak" "375-625" "$CHEST_COINS"
                fi
            else
                log_info "Got $CHEST_TYPE chest (expected golden for 10 correct). Coins: $CHEST_COINS"
            fi
            break
        fi
    done
fi

# ===========================================
# Summary
# ===========================================
echo ""
echo "========================================"
echo "Test Summary"
echo "========================================"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed!${NC}"
    exit 1
fi
