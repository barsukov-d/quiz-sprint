# Daily Challenge Debug Scripts

Scripts for debugging and testing Daily Challenge functionality.

## Quick Reference

```bash
# Show game state and statistics
make daily-debug                    # Your user (1121083057), today
make daily-debug USER=123456789     # Specific user, today
make daily-debug USER=123 DATE=2026-01-27

# Reset for testing
make daily-quick-reset              # Fast reset, no prompts
make daily-reset                    # Interactive reset with confirmation
make daily-reset-all                # Reset all users (be careful!)

# Or run scripts directly
cd backend
./scripts/debug-daily.sh [USER_ID] [DATE]
./scripts/quick-reset.sh [USER_ID] [DATE]
./scripts/reset-daily.sh [USER_ID] [DATE]
```

## Scripts Overview

### 1. `debug-daily.sh` - Inspector Tool

**Shows comprehensive state of Daily Challenge:**

```bash
./scripts/debug-daily.sh                    # Your user, today
./scripts/debug-daily.sh 123456789          # Specific user, today
./scripts/debug-daily.sh 123456789 2026-01-27
```

**Displays:**
- ✅ User's games for the date (all attempts)
- 🔥 Streak history (last 7 days)
- 🏆 Today's leaderboard (top 10)
- 📝 Daily quiz info
- 📊 Statistics (total games, avg score, etc.)
- 🎮 Can play status
- 🛠️ Available actions

**Example output:**
```
╔════════════════════════════════════════╗
║   Daily Challenge Debug Inspector     ║
╚════════════════════════════════════════╝

📅 Games for User 1121083057 on 2026-01-28:
 id | status | streak | score | question_idx | chest_type
----|--------|--------|-------|--------------|------------
 ... | completed | 5 | 920 | 10 | golden

🔥 Streak History (Last 7 days):
 date       | status | streak | score
------------|--------|--------|-------
 2026-01-28 | completed | 5 | 920
 2026-01-27 | completed | 4 | 850
 ...

🏆 Today's Leaderboard (Top 10):
 player_id | score | streak | chest
-----------|-------|--------|-------
 123456    | 1000  | 7      | golden
 ...

✅ User can start Daily Challenge
```

### 2. `quick-reset.sh` - Fast Reset

**Instant reset without prompts** (for rapid testing):

```bash
./scripts/quick-reset.sh                    # Your user, today
./scripts/quick-reset.sh 123456789          # Specific user, today
./scripts/quick-reset.sh 123456789 2026-01-27
```

**Output:**
```
🔄 Quick reset: User 1121083057 on 2026-01-28
✅ Done! Ready to play.
```

**Use case:** Rapid testing cycles (reset → test → reset → test)

### 3. `reset-daily.sh` - Interactive Reset

**Safe reset with confirmation and state preview:**

```bash
./scripts/reset-daily.sh                    # Your user, today
./scripts/reset-daily.sh 123456789          # Specific user, today
./scripts/reset-daily.sh 123456789 2026-01-27
./scripts/reset-daily.sh --all              # All users, today
./scripts/reset-daily.sh --all 2026-01-27   # All users, specific date
./scripts/reset-daily.sh --help             # Show help
```

**Features:**
- Shows current game state before deletion
- Confirmation prompt (prevents accidents)
- Statistics after reset
- Color-coded output
- Support for resetting all users

**Example output:**
```
╔════════════════════════════════════════╗
║   Daily Challenge Reset Tool          ║
╔════════════════════════════════════════╗

Target: User 1121083057 on 2026-01-28

📊 Current State:
 id | status | score
----|--------|-------
 ... | completed | 920

Do you want to proceed with deletion? (y/n)
y

🗑️  Deleting game for user 1121083057 on 2026-01-28
DELETE 1

✅ Reset complete!

📈 Statistics:
 total_games | best_streak
-------------|------------
 5           | 7

🎮 You can now start a fresh Daily Challenge!
```

## Common Use Cases

### Testing First-Time Play
```bash
make daily-quick-reset
# Now test starting fresh Daily Challenge
```

### Testing Retry Functionality
```bash
# 1. Complete a game first (via UI)
# 2. Then retry:
make daily-debug  # Verify completed status
# 3. Test retry button in UI
```

### Testing Streak Mechanics
```bash
# View streak history
make daily-debug

# Reset yesterday to break streak
make daily-reset DATE=2026-01-27

# Or manually set streak in DB:
docker compose -f docker-compose.dev.yml exec -T postgres psql -U quiz_user -d quiz_sprint_dev -c "
UPDATE daily_games
SET current_streak = 10, best_streak = 15
WHERE player_id = '1121083057' AND date = '2026-01-28';
"
```

### Testing Multiple Attempts
```bash
# 1. Complete first attempt
make daily-debug  # See attempt_number = 1

# 2. Don't reset - test retry with coins/ad

# 3. After retry, check both attempts:
make daily-debug  # See attempt_number = 1, 2
```

### Testing Leaderboard
```bash
# Create multiple users' games:
docker compose -f docker-compose.dev.yml exec -T postgres psql -U quiz_user -d quiz_sprint_dev -c "
INSERT INTO daily_games (id, player_id, daily_quiz_id, date, status, session_state, current_streak, best_streak)
VALUES
  (gen_random_uuid(), '111', 'quiz-id', '2026-01-28', 'completed', '{\"base_score\": 1000}', 5, 10),
  (gen_random_uuid(), '222', 'quiz-id', '2026-01-28', 'completed', '{\"base_score\": 950}', 3, 8);
"

make daily-debug  # Check leaderboard
```

### Clear All Data (Nuclear Option)
```bash
# Reset everyone for today
make daily-reset-all

# Or delete all daily games entirely:
docker compose -f docker-compose.dev.yml exec -T postgres psql -U quiz_user -d quiz_sprint_dev -c "
DELETE FROM daily_games;
DELETE FROM daily_quizzes;
"
```

## Direct SQL Commands

### Check if can play
```sql
SELECT COUNT(*) FROM daily_games
WHERE player_id = '1121083057' AND date = CURRENT_DATE;
-- 0 = can play, 1+ = already played
```

### Get current streak
```sql
SELECT current_streak, best_streak, last_played_date
FROM daily_games
WHERE player_id = '1121083057'
ORDER BY date DESC LIMIT 1;
```

### Force complete a game
```sql
UPDATE daily_games
SET status = 'completed',
    session_state = jsonb_set(session_state, '{completed_at}', to_jsonb(extract(epoch from now())::bigint))
WHERE player_id = '1121083057' AND date = CURRENT_DATE;
```

### Add chest reward manually
```sql
UPDATE daily_games
SET chest_type = 'golden',
    chest_coins = 500,
    chest_pvp_tickets = 5,
    chest_bonuses = '["shield", "fifty_fifty"]'::jsonb
WHERE player_id = '1121083057' AND date = CURRENT_DATE;
```

## Troubleshooting

### "No such file or directory"
Make scripts executable:
```bash
chmod +x backend/scripts/*.sh
```

### "docker: command not found"
Make sure Docker is running and you're in the `backend/` directory.

### "relation does not exist"
Database might not be initialized. Run migrations:
```bash
# Check if migrations ran
docker compose -f docker-compose.dev.yml logs postgres | grep "daily_games"
```

### "Connection refused"
Backend services aren't running:
```bash
cd backend
docker compose -f docker-compose.dev.yml up -d
```

## Advanced: Custom Scenarios

### Set specific streak
```bash
docker compose -f docker-compose.dev.yml exec -T postgres psql -U quiz_user -d quiz_sprint_dev << EOF
UPDATE daily_games
SET current_streak = 30, best_streak = 30
WHERE player_id = '1121083057' AND date = CURRENT_DATE;
EOF

make daily-debug  # Verify
```

### Simulate broken streak
```bash
# Delete yesterday's game
make daily-reset DATE=$(date -v-1d +%Y-%m-%d)

# Now today's streak should reset to 1
make daily-debug
```

### Create test leaderboard
```bash
# Script to populate leaderboard with test data
for i in {1..20}; do
  USER_ID="test_$i"
  SCORE=$((RANDOM % 500 + 500))  # Random 500-1000

  docker compose -f docker-compose.dev.yml exec -T postgres psql -U quiz_user -d quiz_sprint_dev -c "
  INSERT INTO daily_games (id, player_id, daily_quiz_id, date, status, session_state, current_streak, best_streak)
  VALUES (gen_random_uuid(), '$USER_ID', 'quiz-id', CURRENT_DATE, 'completed', '{\"base_score\": $SCORE, \"correct_answers\": 8}', 3, 5)
  ON CONFLICT DO NOTHING;
  "
done

make daily-debug  # See populated leaderboard
```

## Tips

- 🚀 Use `make daily-quick-reset` for fastest testing
- 🔍 Use `make daily-debug` to understand current state
- ⚠️ Use `make daily-reset` when you need to see what you're deleting
- 🧹 Use `make daily-reset-all` carefully - affects all users!
- 💾 Always check `daily-debug` before resetting to understand state

## See Also

- `../docs/game_modes/daily_challenge/` - Full Daily Challenge documentation
- `../IMPORT.md` - Quiz import documentation
- `../README.md` - Main backend README
