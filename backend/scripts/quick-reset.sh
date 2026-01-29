#!/bin/bash

# ==========================================
# Quick Daily Challenge Reset (No Prompts)
# ==========================================
# Fast reset for rapid testing cycles

USER_ID=${1:-1121083057}
DATE=${2:-$(date +%Y-%m-%d)}

echo "🔄 Quick reset: User $USER_ID on $DATE"

docker compose -f docker-compose.dev.yml exec -T postgres psql -U quiz_user -d quiz_sprint_dev -c "DELETE FROM daily_games WHERE player_id = '$USER_ID' AND date = '$DATE';" > /dev/null 2>&1

echo "✅ Done! Ready to play."
