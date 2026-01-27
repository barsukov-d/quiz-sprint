<script setup lang="ts">
import { computed } from 'vue'
import type { InternalInfrastructureHttpHandlersLeaderboardEntryDTO } from '@/api/generated'

interface Props {
  leaderboard: InternalInfrastructureHttpHandlersLeaderboardEntryDTO[] | null | undefined
  currentPlayerId?: string
  showRank?: boolean
  maxEntries?: number
}

const props = withDefaults(defineProps<Props>(), {
  showRank: true,
  maxEntries: 10
})

// ===========================
// Computed
// ===========================

const displayedLeaderboard = computed(() => {
  if (!props.leaderboard) return []
  return props.leaderboard.slice(0, props.maxEntries)
})

const getRankBadgeColor = (rank: number) => {
  if (rank === 1) return 'yellow'
  if (rank === 2) return 'gray'
  if (rank === 3) return 'orange'
  return 'blue'
}

const getRankEmoji = (rank: number) => {
  if (rank === 1) return '🥇'
  if (rank === 2) return '🥈'
  if (rank === 3) return '🥉'
  return ''
}

const isCurrentPlayer = (playerId: string) => {
  return playerId === props.currentPlayerId
}
</script>

<template>
  <div class="leaderboard">
    <div class="leaderboard-header">
      <h3 class="text-lg font-semibold flex items-center gap-2">
        <UIcon name="i-heroicons-trophy" class="size-5 text-yellow-500" />
        Top Players
      </h3>
      <p class="text-sm text-gray-500 dark:text-gray-400">Today's daily challenge</p>
    </div>

    <div class="leaderboard-list">
      <div
        v-for="entry in displayedLeaderboard"
        :key="entry.playerId"
        class="leaderboard-entry"
        :class="{ 'current-player': isCurrentPlayer(entry.playerId) }"
      >
        <!-- Rank -->
        <div class="entry-rank">
          <UBadge
            v-if="showRank"
            :color="getRankBadgeColor(entry.rank)"
            size="lg"
            variant="soft"
          >
            <span v-if="entry.rank <= 3" class="text-lg">{{ getRankEmoji(entry.rank) }}</span>
            <span v-else>#{{ entry.rank }}</span>
          </UBadge>
        </div>

        <!-- Avatar & Name -->
        <div class="entry-player">
          <UAvatar
            :src="entry.avatarUrl"
            :alt="entry.username"
            size="md"
          />
          <div class="player-info">
            <p class="player-name">
              {{ entry.username }}
              <UBadge v-if="isCurrentPlayer(entry.playerId)" color="primary" size="xs">
                You
              </UBadge>
            </p>
          </div>
        </div>

        <!-- Score -->
        <div class="entry-score">
          <div class="score-value">
            {{ entry.score }}
          </div>
          <div class="score-label">points</div>
        </div>
      </div>

      <!-- Empty State -->
      <UEmpty
        v-if="displayedLeaderboard.length === 0"
        title="No players yet"
        description="Be the first to complete today's challenge!"
        icon="i-heroicons-user-group"
      />
    </div>
  </div>
</template>

<style scoped>
.leaderboard {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.leaderboard-header {
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgb(var(--color-gray-200));
}

.leaderboard-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.leaderboard-entry {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem;
  border-radius: 0.5rem;
  background: rgb(var(--color-gray-50));
  transition: all 0.2s;
}

.leaderboard-entry:hover {
  background: rgb(var(--color-gray-100));
}

.leaderboard-entry.current-player {
  background: rgb(var(--color-primary-50));
  border: 2px solid rgb(var(--color-primary-200));
}

.entry-rank {
  flex-shrink: 0;
  min-width: 3rem;
  display: flex;
  justify-content: center;
}

.entry-player {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}

.player-info {
  flex: 1;
  min-width: 0;
}

.player-name {
  font-weight: 600;
  color: rgb(var(--color-gray-900));
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.entry-score {
  flex-shrink: 0;
  text-align: right;
}

.score-value {
  font-size: 1.125rem;
  font-weight: 700;
  color: rgb(var(--color-primary-600));
}

.score-label {
  font-size: 0.75rem;
  color: rgb(var(--color-gray-500));
}

/* Dark mode */
@media (prefers-color-scheme: dark) {
  .leaderboard-header {
    border-bottom-color: rgb(var(--color-gray-700));
  }

  .leaderboard-entry {
    background: rgb(var(--color-gray-800));
  }

  .leaderboard-entry:hover {
    background: rgb(var(--color-gray-700));
  }

  .leaderboard-entry.current-player {
    background: rgb(var(--color-primary-900) / 0.3);
    border-color: rgb(var(--color-primary-700));
  }

  .player-name {
    color: rgb(var(--color-gray-100));
  }

  .score-value {
    color: rgb(var(--color-primary-400));
  }
}

/* Mobile optimizations */
@media (max-width: 640px) {
  .leaderboard-entry {
    padding: 0.5rem;
    gap: 0.75rem;
  }

  .entry-rank {
    min-width: 2.5rem;
  }

  .score-value {
    font-size: 1rem;
  }
}
</style>
